APP_NAME=echoprom
CONTAINER_NAME=echoprometheus

# .PHONY: build
# build:
# 	@echo "Building..."
# 	@go build -o bin/$(APP_NAME) main.go

.PHONY: docker-network
docker-network:
	@echo "Creating Docker network..."
	@docker network create monitoring || true

.PHONY: prom-build
prom-build:
	@echo "Building Prometheus docker image..."
	@docker pull prom/prometheus

.PHONY: prom-run
prom-run: docker-network
	@echo "Running Prometheus container..."
	@docker run -d \
		--name $(CONTAINER_NAME) \
		--network monitoring \
        -p 9090:9090 \
        -v $(shell pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
        prom/prometheus --config.file=/etc/prometheus/prometheus.yml

.PHONY: prom-clean
prom-clean:
	@echo "Stopping and removing Prometheus Docker container..."
	@docker stop $(CONTAINER_NAME) || true
	@docker rm $(CONTAINER_NAME) || true

.PHONY: prom-restart
prom-restart:
	@echo "Restarting Prometheus container..."
	@docker restart $(CONTAINER_NAME)

.PHONY: grafana-run
grafana-run:
	@echo "Running Grafana container..."
	@docker run -d \
		--name grafana \
		--network monitoring \
		-p 3000:3000 \
		-e "GF_SECURITY_ADMIN_USER=admin" \
		-e "GF_SECURITY_ADMIN_PASSWORD=admin" \
		grafana/grafana

.PHONY: grafana-clean
grafana-clean:
	@echo "Stopping and removing Grafana container..."
	@docker stop grafana || true
	@docker rm grafana || true

.PHONY: grafana-restart
grafana-restart:
	@echo "Restarting Grafana container..."
	@docker restart grafana

.PHONY: api-run
api-run:
	@echo "Building Dockerfile and Running API..."
	@docker build -t echo-api .
	@docker run -d \
  		--name echo-api \
  		--network monitoring \
  		-p 8081:8081 \
  		echo-api

.PHONY: api-clean
api-clean:
	@echo "Removing API container..."
	@docker stop echo-api || true
	@docker rm echo-api || true

.PHONY: api-restart
api-restart:
	@echo "Restarting API container..."
	@docker restart echo-api

.PHONY: prom-node-run
prom-node-run: prom-restart
	@echo "Running node-exporter container..."
	@docker run -d \
		--net monitoring \
		--name node-exporter \
		--restart unless-stopped \
		-p 9100:9100 \
		-v "/:/host:ro,rslave" \
		quay.io/prometheus/node-exporter:latest \
		--path.rootfs=/etc/host \

.PHONY: prom-node-clean
prom-node-clean:
	@echo "Stopping and removing node-exporter container..."
	@docker stop node-exporter || true
	@docker rm node-exporter || true

.PHONY: prom-node-restart
prom-node-restart:
	@echo "Restarting node-exporter container..."
	@docker restart node-exporter

.PHONY: clean-all
clean-all: prom-clean grafana-clean
	@echo "Removing Docker network..."
	@docker network rm monitoring || true
	@echo "Removing binary..."
	@rm -rf bin/$(APP_NAME)
	@echo "Removing grafana container..."
	@make grafana-clean
	@echo "Removing prometheus container..."
	@make prom-clean
	@echo "Removing api container..."
	@docker stop echo-api || true
	@docker rm echo-api || true

.PHONY: run-all
run-all: prom-build prom-run grafana-run api-run
	@echo "Running..."

.PHONY: clean-all
stop-all: prom-clean grafana-clean api-clean
	@echo "Stopping..."

.PHONY: restart-all
restart-all: prom-restart grafana-restart api-restart
	@echo "Restarting all containers..."
