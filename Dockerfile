FROM ubuntu:22.04 AS a

RUN echo hi > /tmp/hi

FROM ubuntu:22.04

COPY --from=A /tmp/hi /tmp/hi

CMD ["/bin/bash"]
