

echo "🔧 Configurando arquivos de testdata..."

# Detectar estrutura do projeto
if [ -d "pkg/autocomplete" ]; then
    TESTDATA_DIR="pkg/autocomplete/testdata"
elif [ -d "autocomplete" ]; then
    TESTDATA_DIR="autocomplete/testdata"
else
    echo "❌ Diretório do autocomplete não encontrado"
    exit 1
fi

mkdir -p "$TESTDATA_DIR"

# Gerar sample_history.txt
cat > "$TESTDATA_DIR/sample_history.txt" << 'EOF'
go build main.go
go test ./...
go run cmd/main.go
git add .
git commit -m "initial commit"
git push origin main
docker build -t myapp .
docker run -p 8080:8080 myapp
kubectl apply -f deployment.yaml
kubectl get pods
helm install myapp ./chart
terraform init
terraform plan
terraform apply
ansible-playbook playbook.yml
make build
make test
make deploy
npm install
npm run build
npm run test
yarn install
yarn build
python -m pip install -r requirements.txt
python manage.py runserver
python main.py
java -jar app.jar
mvn clean install
gradle build
cargo build
cargo run
rustc main.rs
gcc -o main main.c
./main
ls -la
cd /home/user
mkdir project
rm -rf temp
cp file.txt backup/
mv old.txt new.txt
find . -name "*.go"
grep -r "TODO" .
sed 's/old/new/g' file.txt
awk '{print $1}' data.txt
sort file.txt
uniq file.txt
head -n 10 file.txt
tail -f log.txt
cat /etc/hosts
vim config.yml
nano settings.conf
htop
ps aux
kill -9 1234
systemctl restart nginx
service apache2 start
crontab -e
ssh user@server
scp file.txt user@server:/path/
rsync -av local/ remote/
EOF

# Gerar large_commands.json
cat > "$TESTDATA_DIR/large_commands.json" << 'EOF'
{
  "basic_commands": [
    "help", "exit", "quit", "clear", "history", "pwd", "cd", "ls", "mkdir", "rmdir",
    "touch", "rm", "cp", "mv", "find", "grep", "sed", "awk", "sort", "uniq",
    "head", "tail", "cat", "less", "more", "wc", "diff", "patch", "tar", "gzip"
  ],
  "git_commands": [
    "git init", "git clone", "git add", "git commit", "git push", "git pull",
    "git fetch", "git merge", "git rebase", "git branch", "git checkout",
    "git status", "git log", "git diff", "git stash", "git tag", "git remote",
    "git reset", "git revert", "git cherry-pick", "git bisect", "git blame",
    "git show", "git config", "git clean", "git submodule", "git worktree"
  ],
  "docker_commands": [
    "docker build", "docker run", "docker ps", "docker images", "docker pull",
    "docker push", "docker stop", "docker start", "docker restart", "docker rm",
    "docker rmi", "docker exec", "docker logs", "docker inspect", "docker cp",
    "docker commit", "docker tag", "docker save", "docker load", "docker export",
    "docker import", "docker volume", "docker network", "docker system",
    "docker-compose up", "docker-compose down", "docker-compose build"
  ],
  "kubernetes_commands": [
    "kubectl get", "kubectl describe", "kubectl create", "kubectl apply",
    "kubectl delete", "kubectl edit", "kubectl patch", "kubectl replace",
    "kubectl rollout", "kubectl scale", "kubectl autoscale", "kubectl expose",
    "kubectl port-forward", "kubectl proxy", "kubectl exec", "kubectl logs",
    "kubectl top", "kubectl drain", "kubectl cordon", "kubectl uncordon",
    "kubectl taint", "kubectl label", "kubectl annotate", "kubectl config"
  ],
  "go_commands": [
    "go build", "go run", "go test", "go get", "go mod", "go install",
    "go clean", "go doc", "go fmt", "go generate", "go list", "go version",
    "go env", "go fix", "go tool", "go vet", "go work", "go help"
  ],
  "system_commands": [
    "systemctl start", "systemctl stop", "systemctl restart", "systemctl status",
    "systemctl enable", "systemctl disable", "service start", "service stop",
    "service restart", "service status", "ps aux", "ps -ef", "kill", "killall",
    "pkill", "pgrep", "jobs", "bg", "fg", "nohup", "screen", "tmux",
    "htop", "top", "iotop", "nethogs", "iftop", "df", "du", "free",
    "mount", "umount", "fdisk", "lsblk", "lscpu", "lsusb", "lspci"
  ],
  "network_commands": [
    "ping", "wget", "curl", "ssh", "scp", "rsync", "netstat", "ss",
    "lsof", "iptables", "ufw", "firewall-cmd", "nslookup", "dig",
    "host", "whois", "traceroute", "mtr", "tcpdump", "wireshark",
    "nc", "netcat", "socat", "telnet", "ftp", "sftp"
  ],
  "development_commands": [
    "npm install", "npm run", "npm test", "npm build", "npm start",
    "yarn install", "yarn build", "yarn test", "yarn start", "yarn add",
    "pip install", "pip freeze", "pip list", "python -m", "python3",
    "virtualenv", "conda create", "conda activate", "conda install",
    "mvn clean", "mvn compile", "mvn test", "mvn package", "mvn install",
    "gradle build", "gradle test", "gradle run", "gradle clean",
    "cargo build", "cargo run", "cargo test", "cargo check", "cargo fmt",
    "make", "make install", "make clean", "make test", "cmake", "ninja"
  ],
  "editor_commands": [
    "vim", "nvim", "nano", "emacs", "code", "atom", "sublime",
    "gedit", "kate", "mousepad", "leafpad", "pluma", "xed"
  ],
  "compression_commands": [
    "tar -czf", "tar -xzf", "tar -cjf", "tar -xjf", "zip", "unzip",
    "gzip", "gunzip", "bzip2", "bunzip2", "xz", "unxz", "7z", "rar", "unrar"
  ],
  "monitoring_commands": [
    "tail -f", "watch", "iostat", "vmstat", "sar", "mpstat", "pidstat",
    "strace", "ltrace", "ldd", "objdump", "readelf", "nm", "file", "which",
    "whereis", "locate", "updatedb", "apropos", "man", "info", "help"
  ]
}
EOF

# Gerar arquivo com comandos unicode para testes especiais
cat > "$TESTDATA_DIR/unicode_commands.txt" << 'EOF'
configuração
usuário
contraseña
помощь
αυτόματη
自動完成
🚀start
📁files
⚡fast
🔧config
🌐network
💾save
📊stats
🔍search
⚙️settings
🛠️tools
📝edit
🗂️organize
🔒secure
🌟favorite
EOF

# Gerar arquivo com comandos longos para testes de stress
cat > "$TESTDATA_DIR/long_commands.txt" << 'EOF'
very-long-command-name-that-tests-performance-with-extensive-text-and-many-characters
another-extremely-long-command-with-multiple-hyphens-and-extensive-descriptive-text
super-duper-ultra-mega-long-command-name-for-stress-testing-autocomplete-performance
ridiculously-long-command-name-that-goes-on-and-on-with-many-words-separated-by-hyphens
extraordinarily-verbose-command-name-designed-to-test-the-limits-of-fuzzy-matching-algorithms
EOF

echo "✅ Arquivos de testdata criados em $TESTDATA_DIR:"
echo "   - sample_history.txt (histórico de exemplo)"
echo "   - large_commands.json (comandos categorizados)"
echo "   - unicode_commands.txt (comandos com caracteres especiais)"
echo "   - long_commands.txt (comandos longos para stress test)"
echo ""
echo "💡 Estes arquivos podem ser usados nos benchmarks para testes mais realistas."
