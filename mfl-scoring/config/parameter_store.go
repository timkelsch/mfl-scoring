package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type ParameterStoreStore struct {
	client *ssm.Client
}

func NewParameterStore() *ParameterStoreStore {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}

	client := ssm.NewFromConfig(cfg)
	return &ParameterStoreStore{
		client: client,
	}
}

func (ps *ParameterStoreStore) Find(name string, withDecryption bool) string {
	input := &ssm.GetParameterInput{
		Name:           &name,
		WithDecryption: &withDecryption,
	}

	results, err := ps.client.GetParameter(context.TODO(), input)
	if err != nil {
		panic(err)
	}

	if results.Parameter.Value == nil {
		panic(errors.New(fmt.Sprintf("failed to find parameter %s", name)))
	}

	return *results.Parameter.Value
}
