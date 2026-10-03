
TLC scope directive from the user (2026-10-02): email reporting is forbidden.
Stop all email reporting and its dependency-porting work immediately. Spend no
more cycles on JavaMail, SMTP, MIME/Activation, ImageIO/AWT, image codecs, or JVM
emulation pursued for email. These are excluded from TLC completion criteria,
not remaining requirements. This overrides older mail-related plans and notes.
The user's later instruction authorizes surgical removal of email-only source,
tests, fixtures, resources and stale plans. The main thread owns this removal.
Preserve core distributed behavior, packaged property loading, console output,
generic exceptions, OpenJDK notices and x/text. Resume core TLC parity and port
existing Java tests after implementing their features; do not invent tests.

A rule for storing test vectors and associated test data: do not create a 
directory named testdata. Avoid this as a directory name. Use test_vectors instead. 
The reason is that the Go fuzzer uses the directory name testdata/ for its ephemeral
storage and the directory will get wiped and we could lose the test vectors by 
accident and that would be bad.

Do not backup or rebase in git. Make a corrective additional commit instead. We push
in the background and have seen an accidental git fork and we do not want to lose data.
