TLA+ Exercises for your first steps to winning a Turing award.

Leslie Lamport
last modified on Mon 17 June 2019 at  5:09:56 PST by lamport

------------------------------------------------------------------------
The specification SpanTree contains a spec that's an abstract,
non-distributed version of the algorithm in the paper "An Assertional
Correctness Proof of a Distributed Program", available here:

   http://lamport.azurewebsites.net/pubs/pubs.html#dist

The SpanTree spec bears the same relation to the algorithm in that
paper as the Voting spec bears to the Paxos consensus spec.  Your
problem is to write the TLA+ spec of the distributed algorithm and use
TLC to check that it implements the SpanTree spec.  

The SpanTree spec contains a liveness condition as well as the kind of
safety condition you've seen in the lectures and in the three Paxos
specs.  If you can figure out how to express the appropriate liveness
condition for the distributed algorithm, then you can be quite sure
that your algorithm is correct if TLC shows that it implements the
SpanTree spec for a large enough model.  Otherwise, you will have to
write a safety spec of the distributed algorithm and modify the
SpanTree spec to remove its liveness condition.  You should then check
that TLC finds any error that you introduce in your spec.  (A spec
that allows no behaviors implements any safety spec.)

Remember to first test your spec on really tiny networks.





