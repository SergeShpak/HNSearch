# HNSearch

HNSearch is a utility that extracts data on the [Hacker News](#https://news.ycombinator.com/) queries.

The user can extract data on the queries that were made during a specific time period. In particular, the user can:
1. get the number of distinct queries
2. get the most popular queries

The user can extract data only for the queries stored in a dump file. The current dump contains the requests made between **2015-08-01 00:03:43** and **2015-08-04 05:52:16**.

## General overview

HNSearch consists of two parts: the server and the indexer. The server receives the queries of clients, passes them to the indexer and returns the results to the client. The indexer performs all the search tasks.

### First try

At first, I implemented a straightforward, brute-force solution. The code of the server and the indexer was not separated. Each time the client made a request, the server fetched the required data from the data file parsed and analyzed it.

Here are the exact steps taken by the server:

1. Parse client's request and extract the time period demanded.
2. Retrieve the queries made during the given period from the data file.
3. Parse the retrived data line by line and store the queries in a map. The map's keys are the queries represented as strings, and the values are the number of times the query has been made before.

This algorithm has the complexity __O(n)__, where n are the number of queries made during a certain time period. It is worth mentioning, that the search of the period in the data file has the complexity of __O(log(n))__ (it is a form of a binary search), as the indexer sorts the data file by date, before using it.

This approach works fine with short periods (minutes, hours); however, its performance has a linear dependency on the number of queries made and becomes impractical for larger queries (days, and so forth).

To enhance the performance, I could store the analysed time periods in a cache. I did not want to store the cache in memory, so I chose to continue with the index-based approach.

### Index-based approach

The basic idea was to calculate the results for all the time periods before serving the clients requests. In this case the algorithm runs in a constant time, as serving the requests comes down to looking up the already calculated results.

The indexer builds the necessary search indexes by using the passed data file and stores them in a dedicated folder. The only indexes that we need, are those that contain data on the number of distinct queries or the top queries made during a time period. In this case, counting distinct queries has the complexity of __O(1)__: the indexer reads the number of distinct queries from the file directly. For the time beeing, getting **n** top queries in the current implementation has the complexity of __O(n)__ as the top queries are read line by line from the index file. It seems that the retrieval of the top queries from the index file can be optimized (it is trivial to do it in __O(log(n))__); the difficult part is creating a representation for the user, which demands to loop through the results obtained.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. These instructions explain how to run the service in a Linux/Unix OS.

### Run HNSearch locally

To run the service locally, you need to have the Go environment installed on your machine.

To run HNSearch locally you need to execute the __local-build.sh__ script in the root directory of the repo. It builds the __Server__ and the __Indexer__ applications and fetches the queries data file, if it has not been already done.

To start the server locally run the following command from the root directory of the repo:

```bash
cd server && chmod +x local-run.sh && ./local-run.sh
```

To start the indexer locally open a new terminal and run the following command from the root directory of the repo:

```bash
cd indexer && chmod +x local-run.sh && ./local-run.sh
```

### Run HNSearch in a Kubernetes cluster

**This is an experimental feature and it is not guaranteed to work.**

To run the service in a [Kubernetes](https://kubernetes.io/) cluster, you need to have the [minikube](https://github.com/kubernetes/minikube) installed on your machine.

First, make sure the minikube is configured to work with the docker demon:

```bash
eval $(minikube docker-env)
```

Run the following command from the root directory of the repo:

```bash
chmod +x build.sh && ./build.sh
```

I would build the docker images of the __Server__ and the __Indexer__.

Run them in a kubernetes cluster:

```bash
kubectl create -f k8s/minikube
```

The current implementation of the __Indexer__ is not optimized. It requires quite a bit of RAM for the data file indexing, so creting a standard 1 Gi container would probably not work. That is why the indexer pod configuration file (__./k8s/minikube/indexer__) requires 4 Gi of memory for the pod. If the service is never accessible after running the previous command, check the pod status:

```bash
kubectl get pods
```

If the status of the indexer pod is __Pending__, than it is highly likely, the engine cannot allocate 4 Gi of the memory to the pod.

You can try to run the pod with less memory. To do this change the 4Gi memroy requirement in the indexer pod configuration file. However, that could freeze the kubernetes completely. One of the possible solutions of this problem, would be to restart the minikube service and immidiately delete the indexer pod:

```bash
minikube stop
minikube start
kubectl delete -f k8s/minikube
```

## HNSearch code overview

### General solution

In the current version, the __Server__ and the __Indexer__ are implemented as separated services.

The __./server__ diretory contains the __Server__ code. It is implemented in Go with the use of the [gorilla/mux library](https://github.com/gorilla/mux) for routing. The server parses the cient's query, passes it to the __Indexer__ and returns the results from the indexer to the client.

The __Server__ does not know about the implementation of the __Indexer__, it rather communicates with the __Indexer__ interface. I did it to do a kind of dependency injection. In this case the dependencies are define in the configuration file, and the concrete implementation is choses outside the entity that demands it. This is not a pure DI and the DI from textbooks is hard to implement in Go (one of the main reasons for that is the lack of generics). To my knowledge there is only one  reflection-based DI framework well-known to the go community: [dig](https://github.com/uber-go/dig). However, I decided to use a simpler approach to keep the code flexible and readable at the same time.

I used the http implementation of the __Indexer__: the __Server__ and the __Indexer__ communicate via the HTTP protocol. The __Indexer__ works as has been explained [before](#index-based-approach). It consists of three parts: sorter, parser and engine. The sorter and parser take part in the creating of the indexes.

The sorter sorts the received query data file, as all the subsequent operations are much easier to perform on the sorted list of queries. As the data file can be large, the sorter does the [external sorting](https://en.wikipedia.org/wiki/External_sorting) (sorry for the Wiki reference, could not find anything more solid in the internet; it has some great reference to Knuth, though). The exact size of the data that can fit into the RAM is a configuration file parameter **Buffer** (by default it is 500 MB).

The parser parses the entries in the sorted data file.

The engine adds the passed queries to the index. It also serves the __Server__ requests: when the request is received, the indexer gets the required index file, reads it and sends the results to the __Server__.

### A cherry on top

I decided to make the code of __Indexer__ more general and made two improvements:

1. We can add other data files to the existing indexes. To do this, the data file must be placed into the data directory (see indexer configuration file __./indexer/config.json__). After the restart the __Indexer__ will read the file, find the indexes that need to be recalculated and will add them.

2. We can start to query random time periods. The HTTP __Indexer__ is able to treat such requests out-of-the-box. The __Server__ passes the time period in a POST request, and the passed values are not anyhow constrained. If the indexer can do just one lookup (as in the basic task, it would do that). If it needs to examine multiple indexes, it would read the corresponding index files and aggregate their contents. 

## Running the tests

There are not too many tests for the time being. The most interesting ones are in the __./server/server_test.go__. Basically, they verify that the indexer returns correct results. To run them [run the service locally](#run-hnsearch-locally) and do:

```bash
cd ./server && go test server_test.go -count=1
```

## Project TODOs

This project has a large bottleneck which occurrs during the initial indexes generation. When the indexer processes a time period, it constructs a large map of queries and their count which resides in the RAM. The service may run out of memory on a large period with many requests. There are two things to improve:

1. I should not store the full query in this map. Instead, a dictionary of queries should be created, where every query has a unique ID. In this case every query can be represented by 4 (or 8) bytes, which would decrease the memory consumption.
2. I shou implement indexer sharding. When the index files become critically large, a new inedxer (with a new data store) should be created. In this case, the existing architecture is to be extended with a load balancer, which would route the query from the server to the shard (or shards) that contain the necessary data. In case of multiple shards, we would also need an aggregator to merge the results received from each shard.

Another subject that needs to be researched is the index representation. Today the indexer uses a plain map. However, it is possible that we can do better. Generally, to serve the top queries requests we need an easily mergeable data structure that supports priorities. A good potential candidate are [fibonacci heaps](https://en.wikipedia.org/wiki/Fibonacci_heap). I need to understand if it can be applied to the given problem.

The current implementation has a very poor test coverage. It is another subject to work on.

## Authors

Sergey Shpak

## License

This project is licensed under the terms of the MIT license.
