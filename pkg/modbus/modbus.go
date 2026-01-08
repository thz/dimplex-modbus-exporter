// Copyright 2025 Tobias Hintze
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package modbus

import (
	"context"
	"fmt"
	"time"

	"github.com/simonvetter/modbus"
)

type DataSet struct {
	TemperatureInflow float64
}

type ModbusClient struct {
	client        *modbus.ModbusClient
	serverAddress string
}

func New(addr string) (*ModbusClient, error) {
	m := &ModbusClient{
		serverAddress: addr,
	}

	return m, nil
}

func (m *ModbusClient) Connect() error {
	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     m.serverAddress,
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to create modbus client: %w", err)
	}

	err = client.Open()
	if err != nil {
		return fmt.Errorf("failed to open modbus connection: %w", err)
	}

	m.client = client

	return nil
}

func (m *ModbusClient) RequestPseudoFloat16(ctx context.Context, reg uint16) (float32, error) {
	i16, err := m.RequestInt16(ctx, reg)
	if err != nil {
		return 0, err
	}
	return float32(i16) / 10, nil
}

func (m *ModbusClient) RequestInt16(ctx context.Context, reg uint16) (int16, error) {
	reg16, err := m.client.ReadRegister(reg, modbus.HOLDING_REGISTER)
	if err != nil {
		return 0, fmt.Errorf("failed to read register: %w", err)
	}

	return int16(reg16), nil //nolint:gosec // intentional bit reinterpretation for signed register values
}
