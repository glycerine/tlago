
TLC scope directive from the user (2026-10-02): email reporting is forbidden.
Stop all email reporting and its dependency-porting work immediately. Spend no
more cycles on JavaMail, SMTP, MIME/Activation, ImageIO/AWT, image codecs, or JVM
emulation pursued for email. These are excluded from TLC completion criteria,
not remaining requirements. This overrides older mail-related plans and notes.
The user's later instruction authorizes surgical removal of email-only source,
tests, fixtures, resources and stale plans. The main thread owns this removal.
Preserve core distributed behavior, packaged property loading, console output,
generic exceptions, OpenJDK notices and x/text. Resume core TLC parity and port
existing Java tests after implementing their features; do not invent tests--except
for the new features involved with the distributed checking which is
based on our rpc25519 and tube which of course will require new tests!
We like the BDD driven test first approach especially for new architecture like this.

A rule for storing test vectors and associated test data: do not create a 
directory named testdata. Avoid this as a directory name. Use test_vectors instead. 
The reason is that the Go fuzzer uses the directory name testdata/ for its ephemeral
storage and the directory will get wiped and we could lose the test vectors by 
accident and that would be bad.

Do not backup or rebase in git. Make a corrective additional commit instead. We push
in the background and have seen an accidental git fork and we do not want to lose data.

Testing directive from the user (2026-10-05): do not combine long-running
workloads with -race. Run full long workloads normally, preserving their original
bounds. Use -race only for short, focused concurrency checks so porting can
proceed promptly. Do not use broad race selections that include long tests.

Documentation directive from the user (2026-10-05): keep tlc/HANDOFF.md
readable and current. Use normal spacing between words and numbers, clear prose,
and formatted identifiers. Keep detailed run chronology in PORT_PROGRESS.md
rather than accumulating overlapping status blocks in the handoff.

CLI directive from the user (2026-10-05): use the Java TLC command-line flags
by default. Direct TLC arguments and modelcheck/mc must use the same TLC runner.
Do not restore the --tlc selection flag or the bounded checker's CLI path.
