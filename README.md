# MandatoryActivityFour
Our attempt at ManAct4

System Requirements:

R1 (Spec): Implement a system with a set of nodes and a Critical
Section that represents a sensitive system operation. Any node may
request access to the Critical Section at any time. In this exercise,
the Critical Section can be emulated, for example, by a print statement or
writing to a shared database on the network

R2 (Safety): Only one node may enter the Critical Section at any time

R3 (Liveliness): Every node that requests access to the Critical
Section will eventually gain access

Bulletin Board

- Each node/client holds a queue that gets sorted by lampart clock.

- Communication from each node->node to update their queues

- Who grants access to the first in queue?

- How to use Ricart Agrawala correctly? How do we change states to wanted/is it random?.

- What is a reply in R&A Algorithm?
