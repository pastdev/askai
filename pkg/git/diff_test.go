package git

// From git diff --help:
//
// GENERATING PATCH TEXT WITH -P
//        Running git-diff(1), git-log(1), git-show(1), git-diff-index(1), git-diff-tree(1), or git-diff-files(1) with the -p option produces patch text. You can customize the creation of patch text via the GIT_EXTERNAL_DIFF and the GIT_DIFF_OPTS environment variables (see git(1)), and the
//        diff attribute (see gitattributes(5)).
//
//        What the -p option produces is slightly different from the traditional diff format:
//
//         1. It is preceded by a "git diff" header that looks like this:
//
//                diff --git a/file1 b/file2
//
//            The a/ and b/ filenames are the same unless rename/copy is involved. Especially, even for a creation or a deletion, /dev/null is not used in place of the a/ or b/ filenames.
//
//            When a rename/copy is involved, file1 and file2 show the name of the source file of the rename/copy and the name of the file that the rename/copy produces, respectively.
//
//         2. It is followed by one or more extended header lines:
//
//                old mode <mode>
//                new mode <mode>
//                deleted file mode <mode>
//                new file mode <mode>
//                copy from <path>
//                copy to <path>
//                rename from <path>
//                rename to <path>
//                similarity index <number>
//                dissimilarity index <number>
//                index <hash>..<hash> <mode>
//
//            File modes are printed as 6-digit octal numbers including the file type and file permission bits.
//
//            Path names in extended headers do not include the a/ and b/ prefixes.
//
//            The similarity index is the percentage of unchanged lines, and the dissimilarity index is the percentage of changed lines. It is a rounded down integer, followed by a percent sign. The similarity index value of 100% is thus reserved for two equal files, while 100% dissimilarity
//            means that no line from the old file made it into the new one.
//
//            The index line includes the blob object names before and after the change. The <mode> is included if the file mode does not change; otherwise, separate lines indicate the old and the new mode.
//
//         3. Pathnames with "unusual" characters are quoted as explained for the configuration variable core.quotePath (see git-config(1)).
//
//         4. All the file1 files in the output refer to files before the commit, and all the file2 files refer to files after the commit. It is incorrect to apply each change to each file sequentially. For example, this patch will swap a and b:
//
//                diff --git a/a b/b
//                rename from a
//                rename to b
//                diff --git a/b b/a
//                rename from b
//                rename to a
//
//         5. Hunk headers mention the name of the function to which the hunk applies. See "Defining a custom hunk-header" in gitattributes(5) for details of how to tailor this to specific languages.

// ####################
//      1  diff --git a/README.md b/README.md
//      2  new file mode 100644
//      3  index 0000000..fac08eb
//      4  --- /dev/null
//      5  +++ b/README.md
//      6  @@ -0,0 +1,5 @@
//      7  +# Space in path
//      8  +
//      9  +This scenario introduces a file with space in its name so we can see what a diff
//     10  +looks like with space in the path.
//     11  +
//
// ####################
//      1  diff --git a/file with space.txt b/file with space.txt
//      2  new file mode 100644
//      3  index 0000000..9767b7a
//      4  --- /dev/null
//      5  +++ b/file with space.txt
//      6  @@ -0,0 +1,2 @@
//      7  +This is a test file
//      8  +
//
// ####################
//      1  diff --git a/file with space.txt b/file with space.txt
//      2  index 9767b7a..f1a3e81 100644
//      3  --- a/file with space.txt
//      4  +++ b/file with space.txt
//      5  @@ -1,2 +1,2 @@
//      6  -This is a test file
//      7  +This is a test file with a modification
//      8
//
// ####################
//      1  diff --git a/file with space.txt b/file with space.txt
//      2  deleted file mode 100644
//      3  index f1a3e81..0000000
//      4  --- a/file with space.txt
//      5  +++ /dev/null
//      6  @@ -1,2 +0,0 @@
//      7  -This is a test file with a modification
//      8  -
//
// ####################
//      1  diff --git a/file with space.txt b/file with space.txt
//      2  new file mode 100644
//      3  index 0000000..f187ac1
//      4  --- /dev/null
//      5  +++ b/file with space.txt
//      6  @@ -0,0 +1,2 @@
//      7  +gonna change mode
//      8  +
//
// ####################
//      1  diff --git a/file with space.txt b/file with space.txt
//      2  old mode 100644
//      3  new mode 100755
//
// ####################
//      1  diff --git a/file with space.txt b/file with space.txt
//      2  old mode 100755
//      3  new mode 100644
//      4  index f187ac1..271a4cd
//      5  --- a/file with space.txt
//      6  +++ b/file with space.txt
//      7  @@ -1,2 +1,2 @@
//      8  -gonna change mode
//      9  +gonna change mode and changed text
//     10
//
// ####################
//      1  diff --git a/foo.sh b/foo.sh
//      2  new file mode 100755
//      3  index 0000000..e40a992
//      4  --- /dev/null
//      5  +++ b/foo.sh
//      6  @@ -0,0 +1,4 @@
//      7  +#!/bin/bash
//      8  +
//      9  +echo "foo.sh"
//     10  +
//
// ####################
//      1  diff --git a/foo.sh b/foo.sh
//      2  index e40a992..641f120 100755
//      3  --- a/foo.sh
//      4  +++ b/foo.sh
//      5  @@ -1,4 +1,8 @@
//      6   #!/bin/bash
//      7
//      8  -echo "foo.sh"
//      9  +function main {
//     10  +  echo "foo.sh"
//     11  +}
//     12  +
//     13  +main
//     14
//
// ####################
//      1  diff --git a/foo.sh b/bar.sh
//      2  similarity index 100%
//      3  rename from foo.sh
//      4  rename to bar.sh
//
// ####################
//      1  diff --git a/bar.sh b/foo.sh
//      2  similarity index 55%
//      3  rename from bar.sh
//      4  rename to foo.sh
//      5  index 641f120..110367a 100755
//      6  --- a/bar.sh
//      7  +++ b/foo.sh
//      8  @@ -1,8 +1,8 @@
//      9   #!/bin/bash
//     10
//     11   function main {
//     12  -  echo "foo.sh"
//     13  +  echo "$1"
//     14   }
//     15
//     16  -main
//     17  +main "foo.sh"
//     18

// ltheisen@MM292985-PC ~/egit/pastdev-askai
// $ REBUILD=1 askai cr e43ddd4e1848df08dc0141d0abe8eb544b58878a 0a8a46974a721caa2a0275b442980d26e2a94227 --endpoint grok
// building version 1.0.0-snapshot
// 0 diff --git a/.github/workflows/build.yml b/.github/workflows/build.yml
// 1 index 3f88e36bb5210ab71b19c62a54eca8d093d86bfd..13553f0b646ceb02186649577acc0f72695ed1df 100644
// 2 --- a/.github/workflows/build.yml
// 3 +++ b/.github/workflows/build.yml
// 4 @@ -19,16 +19,9 @@       uses: actions/checkout@v4
// 5      - name: Setup go
// 6        uses: actions/setup-go@v5
// 7        with:
// 8 -        # this version is not the same version as our go.mod specifies because
// 9 -        # the linter fails unless it is more modern:
// 10 -        #   https://github.com/golangci/golangci-lint/issues/5051#issuecomment-2386992469
// 11 -        go-version: '^1.22'
// 12          cache: true
// 13      - name: golangci-lint
// 14 -      uses: golangci/golangci-lint-action@v6
// 15 -      with:
// 16 -        args: -v
// 17 -        version: v1.64.6
// 18 +      uses: golangci/golangci-lint-action@v8
// 19
// 20    test:
// 21      runs-on: ubuntu-latest
// 22 diff --git a/.golangci.yml b/.golangci.yml
// 23 index 10f6df0b79ae80ced6ee52c5b6d3f12e3e978b3d..28f3fe4a6d0c9209f64fa8e59925148670f89ef1 100644
// 24 --- a/.golangci.yml
// 25 +++ b/.golangci.yml
// 26 @@ -1,38 +1,44 @@
// 27 ----
// 28 +version: "2"
// 29  linters:
// 30 -  disable-all: true
// 31 +  default: none
// 32    enable:
// 33 -  - asciicheck
// 34 -  - bidichk
// 35 -  - durationcheck
// 36 -  - exhaustive
// 37 -  - errcheck
// 38 -  - errorlint
// 39 -  - gochecknoinits
// 40 -  - goconst
// 41 -  - gocritic
// 42 -  - gofmt
// 43 -  - gosec
// 44 -  - gosimple
// 45 -  - govet
// 46 -  - ineffassign
// 47 -  - nakedret
// 48 -  - nilerr
// 49 -  - nolintlint
// 50 -  - revive
// 51 -  - sqlclosecheck
// 52 -  - staticcheck
// 53 -  - typecheck
// 54 -  - unparam
// 55 -  - unused
// 56 -  - wastedassign
// 57 -  - whitespace
// 58 -  - wrapcheck
// 59 -issues:
// 60 -  include:
// 61 -  # Reenable some checks golangci-lint disables by default (see golangci-lint run --help)
// 62 -  - EXC0001
// 63 -  - EXC0004
// 64 -  - EXC0005
// 65 -  - EXC0006
// 66 -  - EXC0007
// 67 +    - asciicheck
// 68 +    - bidichk
// 69 +    - durationcheck
// 70 +    - errcheck
// 71 +    - errorlint
// 72 +    - exhaustive
// 73 +    - gochecknoinits
// 74 +    - goconst
// 75 +    - gocritic
// 76 +    - gosec
// 77 +    - govet
// 78 +    - ineffassign
// 79 +    - nakedret
// 80 +    - nilerr
// 81 +    - nolintlint
// 82 +    - revive
// 83 +    - sqlclosecheck
// 84 +    - staticcheck
// 85 +    - unparam
// 86 +    - unused
// 87 +    - wastedassign
// 88 +    - whitespace
// 89 +    - wrapcheck
// 90 +  exclusions:
// 91 +    generated: lax
// 92 +    presets:
// 93 +      - comments
// 94 +    paths:
// 95 +      - third_party$
// 96 +      - builtin$
// 97 +      - examples$
// 98 +formatters:
// 99 +  enable:
// 100 +    - gofmt
// 101 +  exclusions:
// 102 +    generated: lax
// 103 +    paths:
// 104 +      - third_party$
// 105 +      - builtin$
// 106 +      - examples$
// 107 diff --git a/cmd/askai/complete/complete.go b/cmd/askai/complete/complete.go
// 108 index 734241adda5a8ffc3ca097abd95c23dea2d5827f..11a1b5ca52f14df3f86c7c870ccb6429da4f6cbe 100644
// 109 --- a/cmd/askai/complete/complete.go
// 110 +++ b/cmd/askai/complete/complete.go
// 111 @@ -43,6 +43,7 @@                           if err != nil {
// 112                                     return err
// 113                             }
// 114                             if !d.IsDir() {
// 115 +                                   //nolint: gosec // the intent is to include a file from a user supplied location
// 116                                     content, err := os.ReadFile(path)
// 117                                     if err != nil {
// 118                                             return fmt.Errorf("read attachment: %w", err)
// 119 @@ -60,6 +61,7 @@                   if err != nil {
// 120                             return "", fmt.Errorf("attachment walk: %w", err)
// 121                     }
// 122             } else {
// 123 +                   //nolint: gosec // the intent is to include a file from a user supplied location
// 124                     content, err := os.ReadFile(path)
// 125                     if err != nil {
// 126                             return "", fmt.Errorf("read attachment: %w", err)
// 127 diff --git a/cmd/askai/config/config.go b/cmd/askai/config/config.go
// 128 index fdddaa67d56e9059a7ddebb38fd34148abaa42c5..f8ef3b50b420fb20db2d5df7cab15e0447ebe69f 100644
// 129 --- a/cmd/askai/config/config.go
// 130 +++ b/cmd/askai/config/config.go
// 131 @@ -63,9 +63,18 @@ func AddConfig(root *cobra.Command) *Config {
// 132     cfg := Config{
// 133             configSource: cobracfg.ConfigLoader[pkgcfg.Config]{
// 134                     DefaultSources: cfgldr.Sources[pkgcfg.Config]{
// 135 -                           cfgldr.DirSource[pkgcfg.Config]{Path: SystemConfigDir},
// 136 -                           cfgldr.DirSource[pkgcfg.Config]{Path: UserConfigDir},
// 137 -                           cfgldr.DirSource[pkgcfg.Config]{Path: DirectoryConfigDir},
// 138 +                           cfgldr.DirSource[pkgcfg.Config]{
// 139 +                                   Path:      SystemConfigDir,
// 140 +                                   Unmarshal: cfgldr.YamlValueTemplateUnmarshal[pkgcfg.Config](nil),
// 141 +                           },
// 142 +                           cfgldr.DirSource[pkgcfg.Config]{
// 143 +                                   Path:      UserConfigDir,
// 144 +                                   Unmarshal: cfgldr.YamlValueTemplateUnmarshal[pkgcfg.Config](nil),
// 145 +                           },
// 146 +                           cfgldr.DirSource[pkgcfg.Config]{
// 147 +                                   Path:      DirectoryConfigDir,
// 148 +                                   Unmarshal: cfgldr.YamlValueTemplateUnmarshal[pkgcfg.Config](nil),
// 149 +                           },
// 150                     },
// 151             },
// 152     }
// 153 @@ -85,12 +94,12 @@                                 return nil
// 154                     }))
// 155
// 156     cfg.configSource.PersistentFlags(root).FileSourceVarP(
// 157 -           cfgldr.YamlUnmarshal,
// 158 +           cfgldr.YamlUnmarshal[pkgcfg.Config](),
// 159             "config",
// 160             "c",
// 161             "location of one or more config files")
// 162     cfg.configSource.PersistentFlags(root).DirSourceVarP(
// 163 -           cfgldr.YamlUnmarshal,
// 164 +           cfgldr.YamlUnmarshal[pkgcfg.Config](),
// 165             "config-dir",
// 166             "d",
// 167             "location of one or more config directories")
// 168 diff --git a/cmd/askai/root/root.go b/cmd/askai/root/root.go
// 169 index 8d618ff3562c6fecadb4851ba05c10026074d450..37b0a8eda7e1ad9803d93a1602597e058a2e5378 100644
// 170 --- a/cmd/askai/root/root.go
// 171 +++ b/cmd/askai/root/root.go
// 172 @@ -9,6 +9,7 @@     "github.com/pastdev/askai/cmd/askai/models"
// 173     "github.com/pastdev/askai/cmd/askai/tokens"
// 174     "github.com/pastdev/askai/cmd/askai/version"
// 175     "github.com/pastdev/askai/pkg/log"
// 176 +   cfgldrlog "github.com/pastdev/configloader/pkg/log"
// 177     "github.com/spf13/cobra"
// 178  )
// 179
// 180 @@ -23,6 +24,7 @@           //nolint: revive // required to match upstream signature
// 181             PersistentPreRun: func(cmd *cobra.Command, args []string) {
// 182                     log.SetLevel(logLevel)
// 183                     log.SetFormat(logFormat)
// 184 +                   cfgldrlog.Logger = log.Logger
// 185             },
// 186     }
// 187
// 188 diff --git a/go.mod b/go.mod
// 189 index 23f4ffab5a9d328509c72716a75b6c672973036f..56b2437ffeb1c67386eb5096ca1b1d71b25394e9 100644
// 190 --- a/go.mod
// 191 +++ b/go.mod
// 192 @@ -4,7 +4,7 @@ go 1.24.1
// 193
// 194  require (
// 195     dario.cat/mergo v1.0.2
// 196 -   github.com/pastdev/configloader v1.0.1
// 197 +   github.com/pastdev/configloader v1.0.5
// 198     github.com/pastdev/open v1.0.1
// 199     github.com/pkoukk/tiktoken-go v0.1.7
// 200     github.com/sashabaranov/go-openai v1.35.6
// 201 diff --git a/go.sum b/go.sum
// 202 index bc05ea5e9e232c6c12b4cb2348eaf8517b6a7c7d..d9fdedec9f6cf773a6aae525eba698334f4547a5 100644
// 203 --- a/go.sum
// 204 +++ b/go.sum
// 205 @@ -18,8 +18,8 @@ github.com/mattn/go-isatty v0.0.16/go.mod h1:kYGgaQfpe5nmfYZH+SKPsOc2e4SrIfOl2e/yFXSvRLM=
// 206  github.com/mattn/go-isatty v0.0.19/go.mod h1:W+V8PltTTMOvKvAeJH7IuucS94S2C6jfK/D7dTCTo3Y=
// 207  github.com/mattn/go-isatty v0.0.20 h1:xfD0iDuEKnDkl03q4limB+vH+GxLEtL/jb4xVJSWWEY=
// 208  github.com/mattn/go-isatty v0.0.20/go.mod h1:W+V8PltTTMOvKvAeJH7IuucS94S2C6jfK/D7dTCTo3Y=
// 209 -github.com/pastdev/configloader v1.0.1 h1:R6vWBvrXLSq1P5s4emRdG7ejWg4GoK3gZVvCxevd6kQ=
// 210 -github.com/pastdev/configloader v1.0.1/go.mod h1:4gEb04yx46iqDJjqvkhM8P7Fs0DWGIWf3DYlGX03s+8=
// 211 +github.com/pastdev/configloader v1.0.5 h1:fT4pJ7mqpQsqiylj8K0hEPq0DNyTa/ZBkW2XF7fpEHs=
// 212 +github.com/pastdev/configloader v1.0.5/go.mod h1:5etkpu8hsS1WETDR+TMEC2Tk47kQ4VcuMLhMOx4NDQo=
// 213  github.com/pastdev/open v1.0.1 h1:gDUR0KUFBoSvIZmbRy5ucokLwGqp1sluUtrK3PT0ESI=
// 214  github.com/pastdev/open v1.0.1/go.mod h1:BATOny5oMl8pKLTMnkU8qbvw7Z636qMwmnYbqEc2QcM=
// 215  github.com/pkg/errors v0.9.1/go.mod h1:bwawxfHBFNV+L2hUp1rHADufV3IMtnDRdf1r5NINEl0=
//
// [
//   {
//     "diff": "4 @@ -19,16 +19,9 @@       uses: actions/checkout@v4\n5      - name: Setup go\n6        uses: actions/setup-go@v5\n7        with:\n8 -        # this version is not the same version as our go.mod specifies because\n9 -        # the linter fails unless it is more modern:\n10 -        #   https://github.com/golangci/golangci-lint/issues/5051#issuecomment-2386992469\n11 -        go-version: '^1.22'\n12          cache: true\n13      - name: golangci-lint\n14 -      uses: golangci/golangci-lint-action@v6\n15 -      with:\n16 -        args: -v\n17 -        version: v1.64.6\n18 +      uses: golangci/golangci-lint-action@v8\n",
//     "file": ".github/workflows/build.yml",
//     "line_start": 8,
//     "line_end": 11,
//     "suggestion": "Reinstate or update the go-version specification to match the version in go.mod. Removing this could lead to CI builds using a Go version incompatible with the codebase, risking build or test failures. If a specific version is required for linting, document the reason and ensure it aligns with project requirements. [MUST] - Consistency between development and CI environments is critical for reproducibility."
//   },
//   {
//     "diff": "4 @@ -19,16 +19,9 @@       uses: actions/checkout@v4\n5      - name: Setup go\n6        uses: actions/setup-go@v5\n7        with:\n8 -        # this version is not the same version as our go.mod specifies because\n9 -        # the linter fails unless it is more modern:\n10 -        #   https://github.com/golangci/golangci-lint/issues/5051#issuecomment-2386992469\n11 -        go-version: '^1.22'\n12          cache: true\n13      - name: golangci-lint\n14 -      uses: golangci/golangci-lint-action@v6\n15 -      with:\n16 -        args: -v\n17 -        version: v1.64.6\n18 +      uses: golangci/golangci-lint-action@v8\n",
//     "file": ".github/workflows/build.yml",
//     "line_start": 18,
//     "line_end": 18,
//     "suggestion": "Specify a version for golangci-lint-action@v8 or use a specific tag. Using the latest version without pinning can introduce breaking changes unexpectedly in CI. [SHOULD] - Pinning versions ensures stability and predictability in CI workflows."
//   },
//   {
//     "diff": "111 @@ -43,6 +43,7 @@                              if err != nil {\n112                                    return err\n113                                 }\n114                                  if !d.IsDir() {\n115 +                                  //nolint: gosec // the intent is to include a file from a user supplied location\n116                                    content, err := os.ReadFile(path)\n117                                          if err != nil {\n118                                            return fmt.Errorf(\"read attachment: %w\", err)\n",
//     "file": "cmd/askai/complete/complete.go",
//     "line_start": 115,
//     "line_end": 115,
//     "suggestion": "Reevaluate the use of //nolint: gosec for reading user-supplied file paths. While the intent is documented, consider implementing path traversal checks or user input sanitization to prevent security risks like reading sensitive system files. [MUST] - Ignoring security linter warnings without robust mitigation can expose the application to vulnerabilities."
//   },
//   {
//     "diff": "123 @@ -60,6 +61,7 @@                      if err != nil {\n124                            return \"\", fmt.Errorf(\"attachment walk: %w\", err)\n125                      }\n126                  } else {\n127 +                 //nolint: gosec // the intent is to include a file from a user supplied location\n128                    content, err := os.ReadFile(path)\n129                          if err != nil {\n130                            return \"\", fmt.Errorf(\"read attachment: %w\", err)\n",
//     "file": "cmd/askai/complete/complete.go",
//     "line_start": 127,
//     "line_end": 127,
//     "suggestion": "Reevaluate the use of //nolint: gosec for reading user-supplied file paths. Implement checks to prevent path traversal attacks or unauthorized file access. [MUST] - Security risks must be addressed even if the intent is to allow user input."
//   },
//   {
//     "diff": "26 @@ -1,38 +1,44 @@\n27 ----\n28 +version: \"2\"\n29  linters:\n30 -  disable-all: true\n31 +  default: none\n",
//     "file": ".golangci.yml",
//     "line_start": 28,
//     "line_end": 31,
//     "suggestion": "Document the reason for changing from 'disable-all: true' to 'default: none' and adding a 'version' field. If this is to adopt a new configuration format or linting behavior, ensure all team members are aware of the impact on existing code. [COULD] - Clarity in configuration changes helps maintain team alignment and prevents unexpected linting issues."
//   }
// ]
