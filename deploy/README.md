# deploy — dreamtree chain + anchor seam (m3)

systemd units for the v0 devnet on m3. See `../docs/anchoring.md` for the full
runbook. Quick version:

```
make install                 # dreamtreed -> ~/go/bin
go install ./cmd/anchord     # anchord  -> ~/go/bin
sudo cp deploy/*.service /etc/systemd/system/
sudo mkdir -p /etc/dreamtree
echo 'ANCHORD_TOKEN=<secret from infra/.env>' | sudo tee /etc/dreamtree/anchord.env
# copy the initialized chain home to /var/lib/dreamtreed, create the `anchor` key
sudo systemctl daemon-reload
sudo systemctl enable --now dreamtreed anchord
curl localhost:9110/healthz
```

cloudflared ingress: `anchor.dreamtree.org` -> `http://localhost:9110`.
