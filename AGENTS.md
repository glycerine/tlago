
TLC scope directive from the user (2026-10-02): email reporting is forbidden.
Stop all email reporting and its dependency-porting work immediately. Spend no
more cycles on JavaMail, SMTP, MIME/Activation, ImageIO/AWT, image codecs, or JVM
emulation pursued for email. These are excluded from TLC completion criteria,
not remaining requirements. This overrides older mail-related plans and notes.
Make only the minimal integration changes needed to disable email construction
and delivery regardless of email properties or overrides, while preserving core
model loading and normal console output. Resume core TLC parity and its existing
Java tests after implementing their features; do not invent regression/unit tests.

A rule for storing test vectors and associated test data: do not create a 
directory named testdata. Avoid this as a directory name. Use test_vectors instead. 
The reason is that the Go fuzzer uses the directory name testdata/ for its ephemeral
storage and the directory will get wiped and we could lose the test vectors by 
accident and that would be bad.

Do not backup or rebase in git. Make a corrective additional commit instead. We push
in the background and have seen an accidental git fork and we do not want to lose data.
