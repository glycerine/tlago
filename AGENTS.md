
A rule for storing test vectors and associated test data: do not create a 
directory named testdata. Avoid this as a directory name. Use test_vectors instead. 
The reason is that the Go fuzzer uses the directory name testdata/ for its ephemeral
storage and the directory will get wiped and we could lose the test vectors by 
accident and that would be bad.

Do not backup or rebase in git. Make a corrective additional commit instead. We push
in the background and have seen an accidental git fork and we do not want to lose data.
