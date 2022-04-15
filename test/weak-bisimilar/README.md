The examples in this folder are strictly weakly bisimilar, but not strongly bisimilar.

### milner-job-shop

This is a complex example of systems that are weakly bisimilar. 
This example is challenging due to it's non-trivial size. In particular, the second system `milner-7-2.2` has approx. 1.5k states and 5k transitions. And when transformed into an LTS suitable for weakly bisimulation checking it is expanded to 1.5k states and 28k transitions. This presents challenges to the "naive" approach of checking every possible rho with a variation of Cleaveland's algorithm. 

Note that the systems are bisimilar for rho=map[1:1; 2:2; 3:3; 4:4].

### handover

This test case is taken from _Victor, Björn, and Faron Moller. "The Mobility Workbench—a tool for the π-calculus." International Conference on Computer Aided Verification. Springer, Berlin, Heidelberg, 1994._ paper, which in turn takes the formal specifications from _Orava, Fredrik, and Joachim Parrow. "An algebraic verification of a mobile network." Formal aspects of computing 4.6 (1992): 497-543._. The latter in particular specifies the implementation with error handling.

The handover is interesting as its implementation LTS has almost 4k states and 7k transitions. When transformed into a weak LTS, the number of transitions increases to 120k-140k. This allows to test the scalability of the bisimulation algorithm.

There are multiple test cases as follows:
 1. `handover-no-error` - tests that short and easy specification indeed describes the operation of the handover implementation without assuming possibility of an error and thus no internal error handling.
 1. `handover-error-handl` - same as `handover-no-error`, but here the implementation supports error in handover process and has internal error handling as in the original paper by Orava and Parrow. 
 1. `handover-impl` - tests that, as expected from the previous test cases, implementation of handover with error handling and without error handling appear equivalent to an observer.