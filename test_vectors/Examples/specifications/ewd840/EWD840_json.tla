------------------------------- MODULE EWD840_json -------------------------------
EXTENDS EWD840,
        TLC,
        TLCExt, \* Trace operator &
        Json    \* JsonSerialize operator (both in CommunityModules-deps.jar)

(*
  The trick is that TLC evaluates disjunct 'Export' iff 'RealInv' equals FALSE.
  JsonInv is the invariant that we have TLC check, i.e. appears in the config.

  Comment e.g. the "active[i]" conjunct in EWD840!SendMsg to trigger a violation
  of `RealInv`.
*)
JsonInv ==
    \/ RealInv:: Inv \* The ordinary invariant to check in EWD840 module.
    \/ Export:: /\ JsonSerialize("trace.json", Trace)
                /\ TLCSet("exit", TRUE) \* Stop model-checking *without* TLC reporting
                                        \* the usual text-based error trace. Replace
                                        \* with FALSE to also print the error-trace
                                        \* and terminate with non-zero process exit
                                        \* value.

(* 3.
  Grab recent tla2tools.jar and CommunityModules-deps.jar (or Toolbox):
   wget -q https://nightly.tlapl.us/dist/tla2tools.jar \
           https://modules.tlapl.us/releases/latest/download/CommunityModules-deps.jar
*)

=============================================================================