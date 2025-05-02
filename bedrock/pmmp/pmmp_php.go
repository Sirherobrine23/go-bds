package pmmp

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"net/url"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"

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
	phpProgramAndExtension = regex.MustCompile(`(?m)^((?P<Extension>(PHP|EXT)_)?(?P<Program>([a-zA-Z0-9].*))_(VER(SION(S)?)?))$`)
)

// PHP Package info and origin
type PHPSource struct {
	PkgName  string              `json:"name"`     // Package name
	Versions map[string][]string `json:"versions"` // Version source and mirror url
}

// Pocketmine PHP required and tools to build
type PHP struct {
	GitHash    string                  `json:"git_hash"`             // Commit hash to run script
	GitRepo    string                  `json:"git_repo"`             // Git repository url
	GitTag     string                  `json:"git_tag,omitempty"`    // Git repository tag
	PHPVersion string                  `json:"php"`                  // PHP versions
	UnixScript string                  `json:"unix_script"`          // Unix script to build
	WinScript  string                  `json:"win_script,omitempty"` // Windows script
	WinBat     string                  `json:"win_bat,omitempty"`    // Windows script (bat)
	WinOldPs   string                  `json:"win_old_ps,omitempty"` // Windows old script
	WinSh      string                  `json:"win_sh,omitempty"`     // Windows old bash script
	Downloads  map[string]string       `json:"downloads,omitempty"`  // Prebuilds php files
	Tools      map[string][]*PHPSource `json:"tools,omitempty"`      // Tools or extensions to PHP to install/build
}

func gitRepo(repoPath, repoUrl string) (*git.Repository, error) {
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
func (php PHP) checkoutGitRepo(buildPath string) error {
	repo, err := gitRepo(buildPath, php.GitRepo)
	if err != nil {
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
			if err := php.checkoutGitRepo(buildPath); err != nil {
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
			if err := php.checkoutGitRepo(buildPath); err != nil {
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

	// Close jobs if not closed
	defer func() {
		select {
		case <-jobs:
		default:
			close(jobs)
		}
	}()

	for _, buildTarget := range PHPBuilds {
		repoPath, repoUrl := filepath.Join(storageRepo, buildTarget[0]), buildTarget[1]
		repo, err := gitRepo(repoPath, repoUrl)
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

			unixFile, err := commitTree.File("compile.sh")
			if err != nil {
				return err
			} else if UnixCompileContent, err = unixFile.Contents(); err != nil {
				return err
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

var PsExtra = []sh.Value{
	sh.BasicSet{"BUILD_TARGET", "x64"},
	sh.BasicSet{"OPTARG", "x64"},
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

		pkgs := processBuildScript(sh.BashWithValues(Worker.UnixCompileContent, PsExtra))
		lastPhp := js_types.Slice[*PHPSource](pkgs).FindLast(func(input *PHPSource) bool { return input.PkgName == "php" })
		if len(pkgs) == 0 || lastPhp == nil || len(lastPhp.Versions) == 0 || len(lastPhp.Versions) == 0 || slices.Collect(maps.Keys(lastPhp.Versions))[0] == "" {
			continue
		}
		phpInfo.Tools["compile.sh"] = pkgs
		phpInfo.PHPVersion = slices.Collect(maps.Keys(lastPhp.Versions))[0]

		if WinScriptPkgs := processBuildScript(sh.PowershellWithValues(strings.ReplaceAll(Worker.WinScriptContent, "$wc.DownloadFile", "wc.DownloadFile"), PsExtra)); len(WinScriptPkgs) > 0 {
			phpInfo.WinScript = "windows-compile-vs.ps1"
			phpInfo.Tools[phpInfo.WinScript] = WinScriptPkgs
		}
		if WinBatPkgs := processBuildScript(sh.Cmd(Worker.WinBatContent)); len(WinBatPkgs) > 0 {
			phpInfo.WinBat = "windows-compile-vs.bat"
			phpInfo.Tools[phpInfo.WinBat] = WinBatPkgs
		}
		if WinOldPsPkgs := processBuildScript(sh.PowershellWithValues(strings.ReplaceAll(Worker.WinOldPsContent, "$wc.DownloadFile", "wc.DownloadFile"), PsExtra)); len(WinOldPsPkgs) > 0 {
			phpInfo.WinOldPs = "windows-binaries.ps1"
			phpInfo.Tools[phpInfo.WinOldPs] = WinOldPsPkgs
		}
		if WinShPkgs := processBuildScript(sh.BashWithValues(Worker.WinShContent, PsExtra)); len(WinShPkgs) > 0 {
			phpInfo.WinSh = "windows-binaries.sh"
			phpInfo.Tools[phpInfo.WinSh] = WinShPkgs
		}

		*Worker.VersionSlice = append(*Worker.VersionSlice, phpInfo)
	}
}

// Parse php build script
func processBuildScript(scriptInter sh.ProcessSh) []*PHPSource {
	// Extensions and tools to build php to pocketmine
	pkgs := map[string]*PHPSource{}
	for _, vars := range scriptInter.Seq() {
		pkgIndex := slices.IndexFunc(vars, func(v sh.Value) bool { return v.ValueType().IsSet() && phpProgramAndExtension.MatchString(v.KeyName()) })
		if pkgIndex == -1 {
			continue
		}
		pkgVar := vars[pkgIndex]
		pkgName := strings.ToLower(phpProgramAndExtension.FindAllGroup(pkgVar.KeyName())["Program"])
		var pkg *PHPSource
		if pkg = pkgs[pkgName]; pkg == nil {
			pkg = &PHPSource{PkgName: pkgName, Versions: map[string][]string{}}
			pkgs[pkgName] = pkg
		}

		for pkgVersion := range pkgVar.Array() {
			switch pkgVar.KeyName() {
			case "PHP_VERSIONS":
				scriptInter.Back()
				scriptInter.SetKey("PHP_VERSION", pkgVersion)
			}

			for line, lineInfo := range scriptInter.Seq(-1, -5) {
				if !slices.ContainsFunc(lineInfo, func(v sh.Value) bool { return v.ValueType().IsAccess() && v.KeyName() == pkgVar.KeyName() }) {
					continue
				}

				info := []string{}
				fields := js_types.Slice[string](fieldParse(line))
			dw:
				switch fields.At(0) {
				case "get_github_extension", "#get_github_extension":
					info = append(info, fmt.Sprintf("https://github.com/%s/%s/archive/%s%s.tar.gz", fields.At(3), fields.At(4), fields.At(5), fields.At(2)))
				case "download_github_src", "#download_github_src":
					info = append(info, fmt.Sprintf("https://github.com/%s/archive/%s.tar.gz", fields.At(1), fields.At(2)))
				case "download_from_mirror", "#download_from_mirror":
					info = append(info, fmt.Sprintf("https://github.com/pmmp/DependencyMirror/releases/download/mirror/%s", fields.At(1)))
				case "download_file", "#download_file":
					info = append(info, fields.At(1))
				case "get_pecl_extension", "#get_pecl_extension":
					info = append(info, fmt.Sprintf("http://pecl.php.net/get/%s-%s.tgz", fields.At(1), fields.At(2)))
				// $wc.DownloadFile("http://windows.php.net/downloads/releases/php-$PHP_VERSION-Win32-VC14-$target.zip", $tmp_path + "php.zip")
				// case "wc.DownloadFile":
				// 	panic(fmt.Sprint(fields.Slice(1, -1)))
				case "git", "#git":
					if fields.At(1) == "clone" {
						for _, field := range fields {
							if strings.HasPrefix(field, "http") || strings.HasPrefix(field, "git:") || strings.HasPrefix(field, "ssh") {
								info = append(info, field)
								break dw
							}
						}
						break
					}
					for line := range scriptInter.Seq(10, -10) {
						fields = js_types.Slice[string](fieldParse(line))
						if !(fields.At(0) == "git" && fields.At(1) == "clone") {
							continue
						}
						for _, urlStr := range fields.Slice(2, -1) {
							if _, err := url.Parse(urlStr); err != nil {
								continue
							}
							info = append(info, urlStr)
							break dw
						}
					}
				default:
					for _, field := range fields {
						if strings.HasPrefix(field, "http") || strings.HasPrefix(field, "ftp") || strings.HasPrefix(field, "git:") || strings.HasPrefix(field, "ssh") {
							info = append(info, field)
							break dw
						}
					}
				}

				for _, pkgSrc := range info {
					if slices.Contains(pkg.Versions[pkgVersion], pkgSrc) {
						continue
					}
					pkg.Versions[pkgVersion] = append(pkg.Versions[pkgVersion], pkgSrc)
				}
			}
		}
	}

	return slices.Collect(maps.Values(pkgs))
}

func fieldParse(s string) []string {
	s = strings.TrimSpace(s)
	type span struct {
		start int
		end   int
	}
	spans := make([]span, 0, 32)
	start := -1
	for end := 0; end < len(s); end++ {
		rune := rune(s[end])
		switch rune {
		case '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0, '(':
			if start >= 0 {
				spans = append(spans, span{start, end})
				// Set start to a negative value.
				// Note: using -1 here consistently and reproducibly
				// slows down this code by a several percent on amd64.
				start = ^start
			}
		case '"', '\'':
			end++
			if start >= 0 {
				endDouble := strings.IndexRune(s[end:], rune)
				if endDouble == -1 {
					endDouble = len(s[end:])
				}
				spans = append(spans, span{start, end + endDouble})
				end = end + endDouble
			} else {
				endDouble := strings.IndexRune(s[end:], rune)
				if endDouble == -1 {
					endDouble = len(s[end:])
				}
				spans = append(spans, span{end, end + endDouble})
				end = end + endDouble
			}
			start = ^end
		default:
			if start < 0 {
				start = end
			}
		}
	}

	// Last field might end at EOF.
	if start >= 0 {
		spans = append(spans, span{start, len(s)})
	}

	// Create strings from recorded field indices.
	a := make([]string, len(spans))
	for i, span := range spans {
		a[i] = s[span.start:span.end]
	}

	return a
}
