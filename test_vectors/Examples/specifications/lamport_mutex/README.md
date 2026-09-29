# Distributed mutual exclusion
TLA+ specification and proof of Lamport's distributed mutual-exclusion
algorithm that appeared as an example in

L. Lamport:  Time, Clocks and the Ordering of Events in a Distributed
System. CACM 21(7):558-565, 1978.

The module LamportMutex contains the specification of the algorithm,
as well as properties that can be checked using TLC. Since TLC can only
handle finite instances of specifications, you must instantiate the
parameters N (number of processes) and maxClock (bound on the clock
values to be considered by TLC). You must also set ClockConstraint as
a state constraint in the "Advanced Options" tab of the TLC interface.

The module LamportMutex_proofs contains a hierarchical proof of mutual
exclusion, based on the definition of an inductive invariant. It applies
to an arbitrary number of processes, and arbitrary clock values. 
The proof can be checked using TLAPS.

The remaining modules come from the standard proof library. They are
also contained in the TLAPS distribution but are reproduced here so that
the proofs module can be loaded in the Toolbox without generating
syntax errors.