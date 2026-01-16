
git fetch --all --tags
git checkout main
git pull --rebase

echo ">> Building"
GOFLAGS="-p=1 -buildvcs=false" CGO_ENABLED=0 \
go build -trimpath -ldflags "-s -w" -o bin/server ./cmd

echo ">> setcap"
sudo setcap 'cap_net_bind_service=+ep' /app/bin/server

echo ">> Restart service"
sudo systemctl restart goapp

echo ">> Status"
sudo systemctl status goapp --no-pager

echo ">> Limpiando logs del journal"
sudo journalctl --rotate
sudo journalctl --vacuum-time=1s

echo ">> Recent logs"
journalctl -u goapp -n 50 --no-pager | grep -i error
