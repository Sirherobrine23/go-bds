package pmmp

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"net/url"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"unicode"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"sirherobrine23.com.br/go-bds/go-bds/utils/file_checker"
	"sirherobrine23.com.br/go-bds/go-bds/utils/js_types"
	"sirherobrine23.com.br/go-bds/go-bds/utils/regex"
	"sirherobrine23.com.br/go-bds/go-bds/utils/semver"
	"sirherobrine23.com.br/go-bds/go-bds/utils/sh"
	"sirherobrine23.com.br/go-bds/request/v2"
)

var (
	PocketminePHPBuild = "https://github.com/PocketMine/php-build-scripts.git"
	PmmpPHPBuild       = "https://github.com/pmmp/PHP-Binaries.git"
	PHPBuilds          = [][2]string{{"pocketmine", PocketminePHPBuild}, {"pmmp", PmmpPHPBuild}}
)

var (
	// Matchs: ^(<PHP_>|<PHP_>)<Tool>_(VER(SION(S)?)?)
	phpProgramAndExtension = regex.MustCompile(`(?m)^((?P<Extension>(PHP|EXT)_)?(?P<Program>([a-zA-Z_][a-zA-Z0-9_]*))_(VER(SION(S)?)?))$`)
)

// PHP Package info and origin
type PHPSource struct {
	PkgName string   `json:"name"`    // Package name
	Version string   `json:"version"` // Version
	Src     []string `json:"src"`     // Source and mirrors
}

// Pocketmine PHP required and tools to build
type PHP struct {
	GitHash    string                  `json:"git_hash"`    // Commit hash to run script
	GitRepo    string                  `json:"git_repo"`    // Git repository url
	PHPVersion string                  `json:"php"`         // PHP versions
	UnixScript string                  `json:"unix_script"` // Unix script to build
	WinScript  string                  `json:"win_script"`  // Windows script
	WinBat     string                  `json:"win_bat"`     // Windows script (bat)
	WinOldPs   string                  `json:"win_old_ps"`  // Windows old script
	WinSh      string                  `json:"win_sh"`      // Windows old bash script
	Tools      map[string][]*PHPSource `json:"tools"`       // Tools or extensions to PHP to install/build
	Downloads  map[string]string       `json:"downloads"`   // Prebuilds php files
}

// Return semver version from PHP
func (ver PHP) SemverVersion() semver.Version { return semver.New(ver.PHPVersion) }

// Install prebuild binary's
func (php PHP) Install(installPath string) error {
	if urlDownload, ok := php.Downloads[fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)]; ok {
		switch path.Ext(urlDownload) {
		case ".zip":
			return request.Zip(urlDownload, request.ExtractOptions{Cwd: installPath}, nil)
		default:
			return request.Tar(urlDownload, request.ExtractOptions{Cwd: installPath}, nil)
		}
	}
	return fmt.Errorf("prebuild to %s/%s not exists", runtime.GOOS, runtime.GOARCH)
}

// Clone repo and checkout to hash
func (php PHP) checkoutGitRepo(buildPath string, logWrite io.Writer) error {
	repo, err := git.PlainClone(buildPath, false, &git.CloneOptions{URL: php.GitRepo, Progress: logWrite})
	switch err {
	case nil:
	case git.ErrRepositoryAlreadyExists:
		if repo, err = git.PlainOpen(buildPath); err != nil {
			return err
		}
	default:
		return err
	}

	// Get worktree
	work, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("cannot get git Worktree: %s", err)
	}

	// Fetch new changes
	repo.Fetch(&git.FetchOptions{})

	// Checkout to commit
	err = work.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(php.GitHash)})
	if err != nil {
		return fmt.Errorf("cannot checkout: %s", err)
	}
	return nil
}

// Build php localy with Repository scripts
func (php PHP) Build(buildPath string, logWrite io.Writer) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "aix", "plan9", "solaris", "js", "illumos", "ios", "dragonfly":
		return ErrPlatform
	case "windows":
		cmd = exec.Command("powershell", fmt.Sprintf("irm %q | iex", php.WinScript))
		if php.GitRepo != "" {
			if err := php.checkoutGitRepo(buildPath, logWrite); err != nil {
				return err
			}
			if scriptPath := filepath.Join(buildPath, "windows-compile-vs.ps1"); file_checker.IsFile(scriptPath) {
				cmd = exec.Command("powershell", scriptPath)
			} else if scriptPath := filepath.Join(buildPath, "windows-compile-vs.bat"); file_checker.IsFile(scriptPath) {
				cmd = exec.Command("cmd", scriptPath)
			} else if scriptPath := filepath.Join(buildPath, "windows-binaries.ps1"); file_checker.IsFile(scriptPath) {
				cmd = exec.Command("powershell", scriptPath)
			} else if scriptPath := filepath.Join(buildPath, "windows-binaries.sh"); file_checker.IsFile(scriptPath) {
				cmd = exec.Command("powershell", scriptPath)
			} else {
				return errors.New("cannot get valid script build to Windows")
			}
		}
	default:
		cmd = exec.Command("sh", "-c", fmt.Sprintf("curl -Ssl %q | bash -", php.UnixScript))
		if php.GitRepo != "" {
			if err := php.checkoutGitRepo(buildPath, logWrite); err != nil {
				return err
			}
			cmd = exec.Command("bash", filepath.Join(buildPath, "compile.sh"))
		}
	}

	// Check if cmd have error
	if cmd.Err != nil {
		return fmt.Errorf("error on spawn command: %s", cmd.Err)
	}

	cmd.Dir = buildPath
	cmd.Stderr = logWrite
	cmd.Stdout = logWrite
	return cmd.Run()
}

// PHP versions from builds scripts tools
type PHPs []*PHP

func (PHPs) checkout(repoPath, repoUrl string) (*git.Repository, error) {
	repo, err := git.PlainClone(repoPath, false, &git.CloneOptions{URL: repoUrl})
	if err == git.ErrRepositoryAlreadyExists {
		if repo, err = git.PlainOpen(repoPath); err == nil {
			if err := repo.Fetch(&git.FetchOptions{}); err != nil {
				return nil, err
			}
		}
	}
	return repo, err
}

// Process build scripts from Pmmp and Pocketmine php build scripts
func (phpBuildSlice *PHPs) FetchAllScripts(storageRepo string) error {
	// job<-chan workerPayload, wg *sync.WaitGroup
	var wg sync.WaitGroup
	jobs := make(chan workerPayload)

	// Start workers
	for range runtime.NumCPU() * 4 {
		wg.Add(1)
		go commitWorker(jobs, &wg)
	}

	for _, buildTarget := range PHPBuilds {
		repoPath, repoUrl := filepath.Join(storageRepo, buildTarget[0]), buildTarget[1]
		repo, err := phpBuildSlice.checkout(repoPath, repoUrl)
		if err != nil {
			return err
		}

		commits, err := repo.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
		if err != nil {
			return err
		}
		defer commits.Close()
		for {
			commit, err := commits.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			// Get commit tree
			commitTree, err := commit.Tree()
			if err != nil {
				return err
			}

			// Scripts and file Content
			var UnixCompileContent, WinScriptContent, WinBatContent, WinOldPsContent, WinShContent string
			if file, err := commitTree.File("compile.sh"); err == nil {
				if UnixCompileContent, err = file.Contents(); err != nil {
					return err
				}
			}
			if file, err := commitTree.File("windows-compile-vs.ps1"); err == nil {
				WinScriptContent, _ = file.Contents()
			}
			if file, err := commitTree.File("windows-compile-vs.bat"); err == nil {
				WinBatContent, _ = file.Contents()
			}
			if file, err := commitTree.File("windows-binaries.ps1"); err == nil {
				WinOldPsContent, _ = file.Contents()
			}
			// This is garbarge to this scripts,
			// in future before fix sh.Sh i return use this script to get php info
			if file, err := commitTree.File("windows-binaries.sh"); err == nil {
				WinShContent, _ = file.Contents()
			}

			jobs <- workerPayload{
				VersionSlice:       phpBuildSlice,
				gitRepo:            repoUrl,
				gitHash:            commit.Hash.String(),
				UnixCompileContent: UnixCompileContent,
				WinScriptContent:   WinScriptContent,
				WinBatContent:      WinBatContent,
				WinOldPsContent:    WinOldPsContent,
				WinShContent:       WinShContent,
			}
		}
		commits.Close()
	}

	// Done process and wait done process worker loader
	close(jobs)
	wg.Wait()

	// Sort versions
	semver.Sort(*phpBuildSlice)
	return nil
}

type workerPayload struct {
	VersionSlice                                                                                         *PHPs
	gitRepo, gitHash, UnixCompileContent, WinScriptContent, WinBatContent, WinOldPsContent, WinShContent string
}

// Process jobs to new worker load
func commitWorker(job <-chan workerPayload, wg *sync.WaitGroup) {
	defer wg.Done()
	for Worker := range job {
		phpInfo := &PHP{
			GitRepo:    Worker.gitRepo,
			GitHash:    Worker.gitHash,
			Tools:      map[string][]*PHPSource{},
			Downloads:  map[string]string{},
			UnixScript: "compile.sh",
		}

		pkgs := processBuildScript(Worker.UnixCompileContent)
		lastPhp := js_types.Slice[*PHPSource](pkgs).FindLast(func(input *PHPSource) bool { return input.PkgName == "php" })
		if len(pkgs) == 0 || lastPhp == nil {
			continue
		}
		phpInfo.Tools["compile.sh"] = pkgs
		phpInfo.PHPVersion = lastPhp.Version

		if WinScriptPkgs := processBuildScript(Worker.WinScriptContent); len(WinScriptPkgs) > 0 {
			phpInfo.WinScript = "windows-compile-vs.ps1"
			phpInfo.Tools[phpInfo.WinScript] = WinScriptPkgs
		}
		if WinBatPkgs := processBuildScript(Worker.WinBatContent); len(WinBatPkgs) > 0 {
			phpInfo.WinBat = "windows-compile-vs.bat"
			phpInfo.Tools[phpInfo.WinBat] = WinBatPkgs
		}
		if WinOldPsPkgs := processBuildScript(Worker.WinOldPsContent); len(WinOldPsPkgs) > 0 {
			phpInfo.WinOldPs = "windows-binaries.ps1"
			phpInfo.Tools[phpInfo.WinOldPs] = WinOldPsPkgs
		}
		if WinShPkgs := processBuildScript(Worker.WinShContent); len(WinShPkgs) > 0 {
			phpInfo.WinSh = "windows-binaries.sh"
			phpInfo.Tools[phpInfo.WinSh] = WinShPkgs
		}
		*Worker.VersionSlice = append(*Worker.VersionSlice, phpInfo)
	}
}

func processBuildScript(script string) []*PHPSource {
	fileSh := sh.ProcessSh(script)
	scriptLines := fileSh.Lines()
	phpPkgs := []*PHPSource{}
	
	for pkgVar, pkgVersions := range fileSh.Seq() {
		if !phpProgramAndExtension.MatchString(pkgVar) {
			continue
		}
		// "Extension", "Program"
		pkgName, ok := phpProgramAndExtension.FindAllGroup(pkgVar)["Program"]
		if !ok {
			continue
		}
		pkgName = strings.ToLower(pkgName)

		switch pkgVar {
		case "PHP_VERSIONS":
			pkgVar = "PHP_VERSION"
		}

		for lineIndex, line := range scriptLines {
			line = strings.TrimSpace(line)
			if !sh.ProcessSh(line).ContainsVar(pkgVar) {
				continue
			}

			for pkgVersion := range splitVersion(pkgVersions) {
				fileSh := fileSh.Clone()
				fileSh.SetVar(pkgVar, pkgVersion)
				fields := js_types.Slice[string](fieldParse(fileSh.ReplaceWithVar(line)))
				info := &PHPSource{PkgName: pkgName, Version: pkgVersion}

			dw:
				switch fields.At(0) {
				case "get_github_extension", "#get_github_extension":
					info.Src = append(info.Src, fmt.Sprintf("https://github.com/%s/%s/archive/%s%s.tar.gz", fields.At(3), fields.At(4), fields.At(5), fields.At(2)))
				case "download_github_src", "#download_github_src":
					info.Src = append(info.Src, fmt.Sprintf("https://github.com/%s/archive/%s.tar.gz", fields.At(1), fields.At(2)))
				case "download_from_mirror", "#download_from_mirror":
					info.Src = append(info.Src, fmt.Sprintf("https://github.com/pmmp/DependencyMirror/releases/download/mirror/%s", fields.At(1)))
				case "download_file", "#download_file":
					info.Src = append(info.Src, fields.At(1))
				case "get_pecl_extension", "#get_pecl_extension":
					info.Src = append(info.Src, fmt.Sprintf("http://pecl.php.net/get/%s-%s.tgz", fields.At(1), fields.At(2)))
				case "git", "#git":
					if fields.At(1) == "clone" {
						for _, field := range fields {
							if strings.HasPrefix(field, "http") || strings.HasPrefix(field, "git:") || strings.HasPrefix(field, "ssh") {
								info.Src = append(info.Src, field)
								break dw
							}
						}
						break
					}
					for indexBack := 10; indexBack > 0; indexBack-- {
						fields = js_types.Slice[string](fieldParse(fileSh.ReplaceWithVar(scriptLines[lineIndex-indexBack])))
						if !(fields.At(0) == "git" && fields.At(1) == "clone") {
							continue
						}
						for _, urlStr := range fields.Slice(2, -1) {
							if _, err := url.Parse(urlStr); err != nil {
								continue
							}
							info.Src = append(info.Src, urlStr)
							break dw
						}
					}
				default:
					for _, field := range fields {
						if strings.HasPrefix(field, "http") || strings.HasPrefix(field, "ftp") || strings.HasPrefix(field, "git:") || strings.HasPrefix(field, "ssh") {
							info.Src = append(info.Src, field)
							break dw
						}
					}
				}

				if len(info.Src) > 0 {
					phpPkgs = append(phpPkgs, info)
				}
			}
		}
	}

	pkgFilter := []*PHPSource{}
	for _, info := range phpPkgs {
		pkgIndex := slices.IndexFunc(pkgFilter, func(pkg *PHPSource) bool { return pkg.PkgName == info.PkgName })
		if pkgIndex == -1 {
			pkgFilter = append(pkgFilter, info)
			continue
		}
		phpPkgs[pkgIndex].Src = append(phpPkgs[pkgIndex].Src, info.Src...)
	}
	
	for _, info := range pkgFilter {
		for index := range info.Src {
			info.Src[index] = strings.ReplaceAll(info.Src[index], "$OPTARG", "x64")
		}
	}

	return pkgFilter
}

func splitVersion(input string) iter.Seq[string] {
	input = strings.Trim(input, "@() ")
	return func(yield func(string) bool) {
		for input != "" {
			fallback := input
			switch input[0] {
			case '"':
				index := strings.Index(input[1:], "\"")
				if index <= 0 {
					fallback = input[1:]
					input = ""
					break
				}
				fallback = input[1:index]
				input = strings.TrimSpace(input[index+1:])
			case '\'':
				index := strings.Index(input[1:], "'")
				if index <= 0 {
					fallback = input[1:]
					input = ""
					break
				}
				fallback = input[1:index]
				input = strings.TrimSpace(input[index+1:])
			default:
				if !strings.ContainsFunc(input, unicode.IsSpace) {
					input = ""
					break
				}
				index := strings.IndexFunc(input, unicode.IsSpace)
				fallback = input[1:index]
				input = strings.TrimSpace(input[index+1:])
			}

			if !yield(fallback) {
				return
			}
		}
	}
}

func fieldParse(line string) []string {
	fields := strings.Fields(strings.TrimSpace(line))
	for fieldIndex := range fields {
		fields[fieldIndex] = strings.Trim(fields[fieldIndex], `"'`)
	}
	return fields
}
