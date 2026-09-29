# Chang-Roberts algorithm for leader election in a ring
An early algorithm for electing a leader in a unidirectional ring
appeared as

E. Chang, R. Roberts: An improved algorithm for decentralized extrema-finding
in circular configurations of processes. CACM 22(5): 281-283, 1979.

The module contains a PlusCal version of that algorithm. For verifying it,
create an instance such as

N <- 6

Id <- <<27, 4, 42, 15, 63, 9>>

and check the invariants TypeOK and Correctness and the temporal logic
property Liveness.
