package internal

import (
    "os"
    "os/exec"
    "fmt"
    "time"
)

func waitForMySQL(container, rootPass string) error {
	for i := 1; i <= 30; i++ {
		cmd := exec.Command("docker", "exec", container,
			"mysql", "-uroot", fmt.Sprintf("-p%s", rootPass),
			"-e", "SELECT 1;")

		if err := cmd.Run(); err == nil {
			fmt.Println("MySQL is ready.")
			return nil
		}

		fmt.Printf("Waiting for MySQL... (%d/30)\n", i)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("MySQL is not responding in container: %s", container)
}

func grantAllPrivileges(container, rootPass, user string) error {
	sql := fmt.Sprintf(`GRANT ALL PRIVILEGES ON *.* TO '%s'@'%%' WITH GRANT OPTION;`, user)
	cmd := exec.Command("docker", "exec", "-i", container, "mysql", "-uroot", fmt.Sprintf("-p%s", rootPass), "-e", sql)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
