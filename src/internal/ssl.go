package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func ensureRootCA(certsDir string) error {
	rootKey := filepath.Join(certsDir, "rootCA.key")
	rootPem := filepath.Join(certsDir, "rootCA.pem")

	// Если уже есть — пропускаем
	if _, err := os.Stat(rootPem); err == nil {
		return nil
	}

	fmt.Println("Generating rootCA...")
	if err := os.MkdirAll(certsDir, 0755); err != nil {
		return err
	}

	// gen root key
	if err := exec.Command("openssl", "genrsa", "-out", rootKey, "2048").Run(); err != nil {
		return fmt.Errorf("failed to generate rootCA.key: %w", err)
	}

	// gen root cert
	subject := "/C=US/ST=Dev/L=Local/O=DockDev Root/CN=DockDev Root CA"
	cmd := exec.Command("openssl", "req", "-x509", "-new", "-nodes", "-key", rootKey, "-sha256",
		"-days", "3650", "-out", rootPem, "-subj", subject)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate rootCA.pem: %w", err)
	}

	fmt.Println("Importing rootCA.pem into Windows trusted store...")
	importCmd := exec.Command("powershell.exe", "-Command",
		fmt.Sprintf(`Start-Process powershell -Verb runAs -ArgumentList 'certutil -addstore -f Root "%s"'`,
			convertToWindowsPath(rootPem)))
	importCmd.Stdin = os.Stdin
	importCmd.Stdout = os.Stdout
	importCmd.Stderr = os.Stderr
	return importCmd.Run()
}

func generateDomainCert(domain, certsDir string) (string, string, error) {
	domainDir := filepath.Join(certsDir, domain)
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		return "", "", err
	}

	keyPath := filepath.Join(domainDir, domain+".key")
	crtPath := filepath.Join(domainDir, domain+".crt")

	if _, err := os.Stat(crtPath); err == nil {
		return crtPath, keyPath, nil
	}

	csrPath := filepath.Join(domainDir, domain+".csr")
	extPath := filepath.Join(domainDir, domain+".ext")

	// domain key
	if err := exec.Command("openssl", "genrsa", "-out", keyPath, "2048").Run(); err != nil {
		return "", "", err
	}

	// domain CSR
	subject := fmt.Sprintf("/C=US/ST=Dev/L=Local/O=DockDev/CN=%s", domain)
	if err := exec.Command("openssl", "req", "-new", "-key", keyPath, "-out", csrPath, "-subj", subject).Run(); err != nil {
		return "", "", err
	}

	// domain.ext
	extContent := fmt.Sprintf(`authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = %s
`, domain)
	if err := os.WriteFile(extPath, []byte(extContent), 0644); err != nil {
		return "", "", err
	}

	// sign
	cmd := exec.Command("openssl", "x509", "-req", "-in", csrPath,
		"-CA", filepath.Join(certsDir, "rootCA.pem"),
		"-CAkey", filepath.Join(certsDir, "rootCA.key"),
		"-CAcreateserial", "-out", crtPath, "-days", "825", "-sha256", "-extfile", extPath)
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("failed to sign domain certificate: %w", err)
	}

	return crtPath, keyPath, nil
}

func convertToWindowsPath(wslPath string) string {
	cmd := exec.Command("wslpath", "-w", wslPath)
	out, err := cmd.Output()
	if err != nil {
		return wslPath
	}
	return string(out[:len(out)-1])
}
