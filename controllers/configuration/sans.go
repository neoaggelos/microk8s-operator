package configuration

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

var csrConfTemplate = template.Must(template.New("csrConf").Parse(`
[ req ]
default_bits = 2048
prompt = no
default_md = sha256
req_extensions = req_ext
distinguished_name = dn

[ dn ]
C = GB
ST = Canonical
L = Canonical
O = Canonical
OU = Canonical
CN = 127.0.0.1

[ req_ext ]
subjectAltName = @alt_names

[ alt_names ]
{{ range $i, $a := .SANs }}DNS.{{ $i + 1 }} = {{ $a }}
{{ end }}

{{ range $i, $a := .IPs }}IP.{{ $i + 1 }} = {{ $a }}
{{ end }}
#MOREIPS

[ v3_ext ]
authorityKeyIdentifier=keyid,issuer:always
basicConstraints=CA:FALSE
keyUsage=keyEncipherment,dataEncipherment,digitalSignature
extendedKeyUsage=serverAuth,clientAuth
subjectAltName=@alt_names
`))

type templateData struct {
	IPs  []string
	SANs []string
}

func (r *Reconciler) reconcileSANs(ctx context.Context, ips, sans []string) error {
	if len(ips) == 0 && len(sans) == 0 {
		return nil
	}
	ips = append(ips, "127.0.0.1", "10.152.183.1")
	sans = append(sans, "kubernetes", "kubernetes.default", "kubernetes.default.svc", "kubernetes.default.svc.cluster", "kubernetes.default.svc.cluster.local")
	var b bytes.Buffer
	if err := csrConfTemplate.Execute(&b, templateData{IPs: ips, SANs: sans}); err != nil {
		return fmt.Errorf("failed to render csr.conf.template: %w", err)
	}

	updated, err := updateFile(r.CSRConfFile, b.String(), 0660)
	if err != nil {
		return fmt.Errorf("failed to write csr.conf.template: %w", err)
	}
	if !updated {
		log.FromContext(ctx).Info("csr.conf file up to date")
		return nil
	}
	log.FromContext(ctx).Info("updated csr.conf file")
	if err := r.RefreshCertificates(ctx); err != nil {
		return fmt.Errorf("failed to refresh the cluster certificates: %w", err)
	}
	return nil
}
