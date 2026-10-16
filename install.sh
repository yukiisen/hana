mkdir ~/.local/bin -p
curl -fsSL https://github.com/yukiisen/hana/releases/download/0.1/hana-linux -o ~/.local/bin/hana
chmod +x ~/.local/bin/hana
sudo setcap cap_sys_ptrace=eip ~/.local/bin/hana
mkdir ~/.config/hana -p
curl -fsSL https://raw.githubusercontent.com/yukiisen/hana/refs/heads/main/example.config.json -o ~/.config/hana/config.json
