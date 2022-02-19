package chassis

import (
	"context"
	"encoding/json"
	"net/url"
)

type RequestContext struct {
	context.Context
}

func (r RequestContext) UserId() string {
	return r.Value("source-user-id").(string)
}

func (r RequestContext) QueryParam(name string) string {
	params := r.Value("params").(url.Values)
	if params.Has(name) {
		return params.Get(name)
	}
	return ""
}

func (r RequestContext) QueryParamWithDefault(name string, defaultValue string) string {
	params := r.Value("params").(url.Values)
	if params.Has(name) {
		return params.Get(name)
	}
	return defaultValue
}

func (r RequestContext) UrlParam(name string) string {
	params := r.Value("values").(url.Values)
	if params.Has(name) {
		return params.Get(name)
	}
	return ""
}

func (r RequestContext) UrlParamWithDefault(name string, defaultValue string) string {
	params := r.Value("values").(url.Values)
	if params.Has(name) {
		return params.Get(name)
	}
	return defaultValue
}

func (r RequestContext) Payload() []byte {
	return r.Value("payload").([]byte)
}

func (r RequestContext) PayloadType() string {
	return r.Value("payload-type").(string)
}

func (r RequestContext) FromJson(blankItem any) error {
	payload := r.Value("payload").([]byte)
	err := json.Unmarshal(payload, blankItem)
	if err != nil {
		return err
	}
	return nil
}

func NewRequestContext(ctx context.Context) *RequestContext {
	return &RequestContext{
		ctx,
	}
}
