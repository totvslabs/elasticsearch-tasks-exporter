FROM scratch
LABEL maintainer="devops@totvslabs.com"
COPY elasticsearch-tasks-exporter /bin/elasticsearch-tasks-exporter
ENTRYPOINT ["/bin/elasticsearch-tasks-exporter"]
CMD [ "-h" ]
