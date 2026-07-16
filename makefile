.DEFAULT_GOAL := install

.PHONY: install validate-ports

validate-ports:
	@if [ -z "$(strip $(PORTS))" ]; then \
		echo "PORTS is required. Usage: sudo make install PORTS=3306,6379,8080"; \
		exit 1; \
	fi
	@printf '%s\n' "$(PORTS)" | awk -F, '{ \
		for (i = 1; i <= NF; i++) { \
			port = $$i; \
			gsub(/^[[:space:]]+|[[:space:]]+$$/, "", port); \
			if (port !~ /^[0-9]+$$/ || port < 1 || port > 65535) exit 1; \
		} \
	}' || { \
		echo "Invalid PORTS value: use comma-separated ports between 1 and 65535"; \
		exit 1; \
	}

install: validate-ports
	go build -trimpath -ldflags "-w -s" -o ip-inspector && \
  mv ip-inspector /usr/local/sbin/ && \
  sed 's/@PORTS@/$(PORTS)/' ip-inspector.service > /etc/systemd/system/ip-inspector.service && \
  mkdir -p /var/lib/ip-inspector && \
  systemctl daemon-reload && \
  systemctl start ip-inspector && \
  systemctl enable ip-inspector
