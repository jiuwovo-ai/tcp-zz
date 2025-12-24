.PHONY: all build frontend backend clean install

all: build

frontend:
	cd frontend && npm install && npm run build

backend:
	cd backend && go mod tidy && go build -o ../port-forward-dashboard .

build: frontend backend

clean:
	rm -f port-forward-dashboard
	rm -rf backend/static

install: build
	mkdir -p /opt/port-forward-dashboard
	cp port-forward-dashboard /opt/port-forward-dashboard/
	cp -r backend/static /opt/port-forward-dashboard/
	cp port-forward-dashboard.service /etc/systemd/system/
	systemctl daemon-reload
	systemctl enable port-forward-dashboard
	@echo "Installation complete. Run: systemctl start port-forward-dashboard"

uninstall:
	systemctl stop port-forward-dashboard || true
	systemctl disable port-forward-dashboard || true
	rm -f /etc/systemd/system/port-forward-dashboard.service
	rm -rf /opt/port-forward-dashboard
	systemctl daemon-reload
