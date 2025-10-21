package git

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
