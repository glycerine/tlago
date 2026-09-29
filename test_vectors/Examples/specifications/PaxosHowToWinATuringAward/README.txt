TLA+ Exercises for your first steps to winning a Turing award.

Leslie Lamport
last modified on Mon 17 June 2019 at  5:09:56 PST by lamport

------------------------------------------------------------------------
These instructions and the lectures with which they are associated are
provided "as is", without warranty of any kind, express or implied,
including but not limited to the warranty of fitness for the
particular purpose of winning a Turing award.  In no event shall the
author or any persons engaged in the distribution of these
instructions or in the presentation or distribution of videos or other
reproductions of the aforementioned lectures be liable for any claim,
damages or other liability, whether in an action of contract, tort or
otherwise, arising from, out of or in connection with the instructions
or the use or other dealings in these instructions.
-----------------------------------------------------------------------

If you're reading these instructions, you should have attended or
watched the videos of my lectures titled "The Paxos Algorithm - or How
to Win a Turing Award".  This file contains instructions for follow-up
exercises.

First, you should install the Toolbox on your computer and run it.
The instructions for doing that or in the "Obtaining the Toolbox"
section of this Web page

   https://lamport.azurewebsites.net/tla/toolbox.html?unhideBut=hide-obtain&unhideDiv=obtain

except that you should download the Toolbox from here

   https://nightly.tlapl.us/products/

instead of from either of the sites on that Web page.  You will
thereby obtain a pre-release version of the Toolbox containing many
new--and, since this is less than exhaustively tested software,
exciting--features.

The .zip file containing this file also contains files for the three
specifications described in my lecture: Consensus, Voting, and Paxos.
Each specification consists of a .tla file and a .toolbox directory
containing one or more models with which to run the TLC model checker
on those specs.  For each of those .tla files, you can create a spec
in the Toolbox by clicking on "File > Open Spec > Add New Spec..." and
then clicking on "Browse..." and selecting the appropriate .tla file
to be the root module.  Use the default "Specification name" and leave
"Import existing models" checked.

You should view the specs in order: Consensus, Voting, Paxos.  You
should read the TLA+ source files.  Pretty-printed pdf files of the
specs are included; comparing the source file with the pretty-printed
version might help you understand some TLA+ notation.  

You probably didn't completely follow the specs when presented in the
lectures.  The source files have plenty of comments to help you fully
understand them now.  You can't win a Turing award by being satisfied
with a vague understanding of algorithms.

At the end of the specs are descriptions of the TLC models that are
included, instructions for running TLC, some small exercises to help
you become familiar with TLC, and a couple of exercises to help you
understand the algorithms.

----

After you understand the algorithms and are familiar with using the
Toolbox and TLC, you should choose an algorithm that you find
interesting, write a TLA+ spec of it, and check your spec with TLC.
Use the simplest possible algorithm that interests you.  If you have
trouble describing the algorithm in TLA+, look for a TLA+ spec of a
similar kind of algorithm.  The Toolbox's Welcome page contains a link
to the GitHub collection of TLA+ examples.  There are also a few specs
that you can open by clicking on buttons on that page.  (The Welcome
page is shown when you run the Toolbox for the first time and whenever
you close the currently open spec.)

You will know that you've written a correct spec if TLC reports that

  - it doesn't satisfy a property you asked it to check and 

  - examining its error trace reveals an error in the algorithm 
    rather than an error in the spec.  

If TLC finds no error, it means that either the algorithm is correct
or your spec isn't describing the algorithm you think it is.  The best
way to check which is the case is to introduce a small modification in
the spec that should make it describe an incorrect version of the
algorithm, and then see if TLC finds the error.

-----

If you can't think of an appropriate, check out the Spanning Tree example.





