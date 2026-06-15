#!/usr/bin/env bash
# One-time setup script for the sumpfgottheit/packages GitHub Pages repo.
# Run this once locally, then push the result and configure the GitHub Secrets.
#
# Usage: ./scripts/init-packages-repo.sh <path-to-cloned-packages-repo>
set -euo pipefail

PACKAGES_REPO="${1:?Usage: $0 <path-to-packages-repo>}"
PACKAGES_REPO="$(realpath "$PACKAGES_REPO")"

for cmd in gpg; do
  command -v "$cmd" >/dev/null 2>&1 || { echo "Error: $cmd not found — install it first"; exit 1; }
done

echo "==> Generating GPG signing key (no passphrase, for CI automation)"
GPG_PARAMS="$(mktemp)"
cat > "$GPG_PARAMS" <<EOF
%no-protection
Key-Type: RSA
Key-Length: 4096
Subkey-Type: RSA
Subkey-Length: 4096
Name-Real: sumpfgottheit packages
Name-Email: packages@sumpfgottheit.github.io
Expire-Date: 0
%commit
EOF
gpg --batch --gen-key "$GPG_PARAMS"
rm "$GPG_PARAMS"

FINGERPRINT=$(gpg --list-keys --with-colons 'packages@sumpfgottheit.github.io' \
  | awk -F: '/^fpr/{print $10; exit}')
echo "    Key fingerprint: $FINGERPRINT"

echo "==> Exporting public key"
gpg --armor --export "$FINGERPRINT" > "$PACKAGES_REPO/gpg.key"

echo "==> Creating GitHub Pages files"
touch "$PACKAGES_REPO/.nojekyll"
cat > "$PACKAGES_REPO/index.html" <<'HTMLEOF'
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>chooz packages</title>
  <style>
    body { font-family: monospace; max-width: 800px; margin: 2rem auto; padding: 0 1rem; }
    pre  { background: #f4f4f4; padding: 1rem; overflow-x: auto; }
  </style>
</head>
<body>
<h1>chooz package repository</h1>

<h2>Rocky Linux 9 / RHEL-compatible</h2>
<pre>
sudo rpm --import https://sumpfgottheit.github.io/packages/gpg.key

sudo tee /etc/yum.repos.d/chooz.repo &lt;&lt;'EOF'
[chooz]
name=chooz packages
baseurl=https://cloud.sumpfgottheit.casa/rpms/rocky/9/$basearch/
enabled=1
gpgcheck=1
gpgkey=https://sumpfgottheit.github.io/packages/gpg.key
EOF

sudo dnf install chooz
</pre>
</body>
</html>
HTMLEOF

echo ""
echo "============================================================"
echo "  SECRETS — add these to github.com/sumpfgottheit/chooz"
echo "  Settings → Secrets and variables → Actions → Secrets"
echo "============================================================"
echo ""
echo "Secret name : PACKAGES_GPG_KEY"
echo "Secret value (copy everything, including the BEGIN/END lines):"
echo "---"
gpg --armor --export-secret-keys "$FINGERPRINT"
echo "---"
echo ""
echo "============================================================"
echo "  NEXT STEPS"
echo "============================================================"
echo "1.  cd $PACKAGES_REPO"
echo "2.  git add -A && git commit -m 'chore: init package repository'"
echo "3.  git push"
echo "4.  On GitHub: Settings → Pages → Deploy from branch → main / (root)"
echo "5.  Add the secret above to sumpfgottheit/chooz"
