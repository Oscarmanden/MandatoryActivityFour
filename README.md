# MandatoryActivityFour
How to run a demo:  
Open up the project from 3 different terminal's. CD into the nodes folder and type the following:  
Terminal 1: $go run . -id=1 -port=:50051  
Terminal 2: $go run . -id=2 -port=:50052  
Terminal 3: $go run . -id=3 -port=:50053  
  
The node will wait 10 seconds and then try to gain access to the critical section by calling RequestCS()
