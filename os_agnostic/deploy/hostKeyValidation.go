package deploy

/*

	certCheck := &ssh.CertChecker{
		IsHostAuthority: hostAuthCallback(),
		IsRevoked:       certCallback(s),
		HostKeyFallback: hostCallback(s),
	}

func hostAuthCallback() HostAuthorityCallBack {
	return func(p ssh.PublicKey, addr string) bool {
		return true
	}
}

func certCallback(s *SSHServer) IsRevokedCallback {
	return func(cert *ssh.Certificate) bool {
		s.Cert = *cert
		s.IsSSH = true

		return false
	}
}

func hostCallback(s *SSHServer) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		s.Hostname = hostname
		s.PublicKey = key
		return nil
	}
}

*/