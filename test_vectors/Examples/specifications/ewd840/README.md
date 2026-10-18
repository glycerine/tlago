# Termination detection in a ring
A specification of Dijkstra's algorithm for termination detection
in a ring. The algorithm was published as
Edsger W. Dijkstra: Derivation of a termination detection algorithm for distributed computations. Inf. Proc. Letters 16:217-219 (1983).
It appears as [EWD 840](https://www.cs.utexas.edu/users/EWD/ewd08xx/EWD840.PDF).

The module EWD840 contains the specification of the algorithm in TLA+, as well
as its correctness properties (and some non-properties) that can be model checked
using TLC. You'll need to instantiate the parameter N by a concrete value, say, 4.

The module EWD840_proofs contains hierarchical proofs of type correctness and of
the main safety property. These proofs can be checked using TLAPS.

The module SyncTerminationDetection contains a high-level specification of the problem that is implemented by Dijkstra's algorithm.
