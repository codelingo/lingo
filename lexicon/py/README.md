# source code for python lexicon. 

## To build:
$ docker build -t codelingo/py .

## To run:
$ docker run -it --net=host codelingo/lexicon-py ./script/server

Based on this example: https://github.com/grpc/grpc-docker-library/blob/master/0.10/python/README.md