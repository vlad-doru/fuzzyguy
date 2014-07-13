Welcome to FuzzyGuy!
=====================

**FuzzyGuy** is an open-source project aimed at creating a performant and
scalable service for fuzzy searching. The backend service may be extremely
useful to implement facilities such as *Did you mean* at a low cost.

Development
---------
**FuzzyGuy** repository consists of a back-end service developed mainly in *Go*,
which only uses *Python* to generate test files useful for benchmarking,
and a sample application developed in *Python* which interacts with the daemon
via a RESTful API.

Documentation
---------
Documentation about the service may be found in this project's wiki where you
can find how you can interact with the RESTful API and useful information about
the project.

Running a demo
---------
In order to build and run the service you need to have Go installed. You can
then run `./server.sh start` to start it and `./server.sh stop` to stop it.

To run the demo app you can go into *demoapp* directory and run `./server.py`.
You can then access the app at *localhost:5000* in your browser.
