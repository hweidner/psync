Contributing to psync
=====================

Filing issues
-------------

To file a bug or discuss new features, an issue can be created. If you are unsure
if the behavior you found is a bug, or if the issue is security-related, please
contact me by e-mail first.

Please open a new issue for each new problem or proposal. Don't collect multiple
unrelated topics in one issue.

When filing a bug, please make sure you include all the information to
reproduce it:

1. What version of psync are you using (version number or commit ID)?
2. What version of Go did you use to compile psync (`go version`)?
3. What operating system and processor archtecture are you using?
4. What filesystems and mount/export options are you using, separate
   for the source and destination file system?
5. What did you do?
6. What did you expect?
7. What happened?

Contributing code
-----------------

If you believe you can fix a bug or write code for a new feature, feel free to
contribute. However, before you submit source code or a pull request, please
consider the following guidelines.

### Discuss before submission

Do not submit code unless the change has been discussed, either publicly in
the issues, or with me personally. There must be consent on what should be
done, and at least roughly, how it will be done.

As an exception, if the code fixes an issue already listed as a TODO in the
README.md or a `TODO` mark in the source code, it is likely to be accepted
(if the other guidelines are met).

### Fix issues one by one

Each pull request should fix exactly one issue. If a submission contains fixes
to several unrelated issues, it will be rejected if one of them is to be
rejected, even if the others are fine.

Note that Github always cares about the Git commit history to
[tell a story](https://github.com/git-guides/git-commit). It looks odd when
commits are accepted and parts of them get later reverted.

### Respect the direction of the project

Although psync is currently more of a replacement to "cp -r", the goal of the
project is to be a (very much stripped down) parallel variant of rsync. Patches
that leave this road are less likely (although not impossible) to be accepted.

See [GOALS.md](GOALS.md) for a rough plan of the project goals.

### Avoid incompatible changes

Although a 0.x version does not have to be strictly backwards compatible, it
is annoying for users when new versions of a tool change the behaviour,
especially when they have to change their scripts. Try to avoid incompatible
changes, unless the result of the discussion is that they are necessary.

As a conseqence, don't change the naming or meaning of existing command line
options. Introduce new options only for features that are expected to stay.

### Don't inject new dependencies

The source code does currently only import packages from the Go standard
library. Changes that inject dependencies to third party packages must be
discussed first; I will accept them only if they are really advantageous to
the project and there is no proper way to implement them without these
dependencies. The discussion must include a licence compatibility check.

Patches that require CGO (e.g. calls to C code) or contain assembler code
will be rejected.

If you change requires the `unsafe` package, prior discussion is required.
If unsafe code can't be avoided, it must be put in a very short function
within an own source code file. The `unsafe` package may only be imported
there.

### Don't lock out older systems

Please do not submit code that relies on features from very late Go releases,
unless there are compelling reasons to do so. psync is used for enterprise
storage migrations. In such environments, stable and long term release
operating systems can be found. As a rule of thumb, the code should compile
even with the Go version shipped with Debian stable and the backports to
Debian oldstable. (At the time of writing, this is Go 1.11).

This does also mean that the psync projected is not being migrated to
[Go modules](https://blog.golang.org/using-go-modules) support yet.

### Keep the documentation in sync

If you code changes require updates to the documentation, keep in mind that
the `README.md` and `go.doc` files must be kept in sync. If the code also
changes the command line interface, check if the `usage()` function in
`psync.go` has to be adapted.

