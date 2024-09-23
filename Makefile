APP_NAME=echoprom
PROM_CONTAINER_NAME=echoprometheus
GRAFANA_CONTAINER_NAME=grafana


## GRAFANA BACKUP
BACKUP_DIR=$(HOME)/backups/grafana
TIMESTAMP=$(shell date +%Y%m%d%H%M%S)
BACKUP_PATH=$(BACKUP_DIR)/grafana_backup_$(TIMESTAMP)

# Criando diretório de backup
.PHONY: create-backup-dir
create-backup-dir:
	@echo "Creating backup directory..."
	@mkdir -p $(BACKUP_PATH)

# Fazendo backup do banco de dados do Grafana
.PHONY: grafana-backup
grafana-backup: create-backup-dir
	@echo "Backing up Grafana database..."
	@docker cp grafana:/var/lib/grafana/grafana.db $(BACKUP_PATH)/grafana.db
	@docker cp grafana:/var/lib/grafana/plugins $(BACKUP_PATH)/plugins
	@docker cp grafana:/etc/grafana/grafana.ini $(BACKUP_PATH)/grafana.ini 2>/dev/null || :
	@docker cp grafana:/etc/grafana/provisioning $(BACKUP_PATH)/provisioning 2>/dev/null || :
	@echo "Backup saved at $(BACKUP_PATH)"

.PHONY: grafana-restore
grafana-restore:
	@if [ -d "$(BACKUP_DIR)" ]; then \
		echo "Restoring Grafana database..."; \
		docker stop grafana; \
		docker cp $(BACKUP_PATH)/grafana.db grafana:/var/lib/grafana/grafana.db; \
		docker cp -r $(BACKUP_PATH)/plugins grafana:/var/lib/grafana/plugins; \
		docker cp $(BACKUP_PATH)/grafana.ini grafana:/etc/grafana/grafana.ini 2>/dev/null || :; \
		docker cp $(BACKUP_PATH)/provisioning grafana:/etc/grafana/provisioning 2>/dev/null || :; \
		docker start grafana; \
		echo "Grafana database restored"; \
	else \
		echo "No Grafana backup to restore"; \
	fi

## Prometheus backup
PROM_BACKUP_DIR=$(HOME)/backups/prometheus
PROM_BACKUP_PATH=$(PROM_BACKUP_DIR)/prometheus_backup_$(TIMESTAMP)

# Criando diretório de backup
.PHONY: create-prom-backup-dir
create-prom-backup-dir:
	@echo "Creating backup directory..."
	@mkdir -p $(PROM_BACKUP_PATH)

# Fazendo backup do banco de dados do Prometheus
.PHONY: prometheus-backup
prom-backup: create-prom-backup-dir
	@echo "Backing up Prometheus data..."
	@docker cp $(PROM_CONTAINER_NAME):/prometheus $(PROM_BACKUP_PATH)/prometheus
	@echo "Backup saved at $(PROM_BACKUP_PATH)"

.PHONY: prom-restore
prom-restore:
	@if [ -d "$(PROM_BACKUP_DIR)" ]; then \
		echo "Restoring Prometheus data..."; \
		docker stop $(PROM_CONTAINER_NAME); \
		docker cp $(PROM_BACKUP_PATH)/prometheus prometheus:/prometheus; \
		docker start $(PROM_CONTAINER_NAME); \
		echo "Prometheus data restored"; \
	else \
		echo "No Prometheus backup to restore"; \
	fi

.PHONY: backup-all
backup-all: grafana-backup prom-backup
	@echo "Backup all..."

.PHONY: restore-all
restore-all: grafana-restore prom-restore
	@echo "Restore all..."

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
		--name $(PROM_CONTAINER_NAME) \
		--network monitoring \
        -p 9090:9090 \
        -v $(shell pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
        prom/prometheus --config.file=/etc/prometheus/prometheus.yml

.PHONY: prom-clean
prom-clean: prom-backup
	@echo "Stopping and removing Prometheus Docker container..."
	@docker stop $(PROM_CONTAINER_NAME) || true
	@docker rm $(PROM_CONTAINER_NAME) || true

.PHONY: prom-restart
prom-restart:
	@echo "Restarting Prometheus container..."
	@docker restart $(PROM_CONTAINER_NAME)

.PHONY: grafana-run
grafana-run:
	@echo "Running Grafana container..."
	@docker run -d \
		--name $(GRAFANA_CONTAINER_NAME) \
		--network monitoring \
		-p 3000:3000 \
		-e "GF_SECURITY_ADMIN_USER=admin" \
		-e "GF_SECURITY_ADMIN_PASSWORD=admin" \
		grafana/grafana

.PHONY: grafana-clean
grafana-clean: grafana-backup
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
clean-all: backup-all prom-clean grafana-clean api-clean prom-node-clean
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
run-all: prom-build prom-run grafana-run api-run prom-node-run restore-all
	@echo "Running..."

.PHONY: clean-all
stop-all: prom-clean grafana-clean api-clean prom-node-clean
	@echo "Stopping..."

.PHONY: restart-all
restart-all: backup-all prom-restart grafana-restart api-restart prom-node-restart
	@echo "Restarting all containers..."

.PHONY: rebuild-restart-all
rebuild-restart-all: clean-all run-all
	@echo "Rebuilding and restarting all containers..."