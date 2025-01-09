install:
	go build -trimpath -ldflags "-w -s" -o ip-inspector && \
  mv ip-inspector /usr/local/sbin/ && \
  cp ip-inspector.service /etc/systemd/system/ && \
  mkdir -p /var/lib/ip-inspector && \
  systemctl daemon-reload && \
  systemctl start ip-inspector && \
  systemctl enable ip-inspector
