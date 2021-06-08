package app_logs

type Providers map[string]Provider

func (p Providers) Find(logProvider string) Provider {
	if logProvider, ok := p[logProvider]; ok {
		return logProvider
	}
	return nil
}
