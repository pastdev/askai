package git

// ####################
// diff --git a/README.md b/README.md
// new file mode 100644
// index 0000000..fac08eb
// --- /dev/null
// +++ b/README.md
// @@ -0,0 +1,5 @@
// +# Space in path
// +
// +This scenario introduces a file with space in its name so we can see what a diff
// +looks like with space in the path.
// +
//
// ####################
// diff --git a/file with space.txt b/file with space.txt
// new file mode 100644
// index 0000000..9767b7a
// --- /dev/null
// +++ b/file with space.txt
// @@ -0,0 +1,2 @@
// +This is a test file
// +
//
// ####################
// diff --git a/file with space.txt b/file with space.txt
// index 9767b7a..f1a3e81 100644
// --- a/file with space.txt
// +++ b/file with space.txt
// @@ -1,2 +1,2 @@
// -This is a test file
// +This is a test file with a modification
//
//
// ####################
// diff --git a/file with space.txt b/file with space.txt
// deleted file mode 100644
// index f1a3e81..0000000
// --- a/file with space.txt
// +++ /dev/null
// @@ -1,2 +0,0 @@
// -This is a test file with a modification
// -
//
// ####################
// diff --git a/file with space.txt b/file with space.txt
// new file mode 100644
// index 0000000..f187ac1
// --- /dev/null
// +++ b/file with space.txt
// @@ -0,0 +1,2 @@
// +gonna change mode
// +
//
// ####################
// diff --git a/file with space.txt b/file with space.txt
// old mode 100644
// new mode 100755
//
// ####################
// diff --git a/file with space.txt b/file with space.txt
// old mode 100755
// new mode 100644
// index f187ac1..271a4cd
// --- a/file with space.txt
// +++ b/file with space.txt
// @@ -1,2 +1,2 @@
// -gonna change mode
// +gonna change mode and changed text
//
// ####################
// diff --git a/foo.sh b/foo.sh
// index e40a992..641f120 100755
// --- a/foo.sh
// +++ b/foo.sh
// @@ -1,4 +1,8 @@
//  #!/bin/bash
//
// -echo "foo.sh"
// +function main {
// +  echo "foo.sh"
// +}
// +
// +main
