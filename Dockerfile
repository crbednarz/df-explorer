FROM ubuntu:latest

RUN echo hi > /tmp/hi
CMD ["/bin/bash"]
