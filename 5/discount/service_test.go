package discount

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePercent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		percent int
		wantErr error
	}{
		{
			name:    "zero_percent_is_valid",
			percent: 0,
			wantErr: nil,
		},
		{
			name:    "ten_percent_is_valid",
			percent: 10,
			wantErr: nil,
		},
		{
			name:    "hundred_percent_is_valid",
			percent: 100,
			wantErr: nil,
		},
		{
			name:    "negative_percent_returns_error",
			percent: -1,
			wantErr: ErrInvalidPercent,
		},
		{
			name:    "above_hundred_percent_returns_error",
			percent: 101,
			wantErr: ErrInvalidPercent,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidatePercent(tt.percent)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestApply(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		price     int
		percent   int
		wantPrice int
		wantErr   error
	}{
		{
			name:      "regular_discount",
			price:     1000,
			percent:   20,
			wantPrice: 800,
			wantErr:   nil,
		},
		{
			name:      "zero_percent_discount",
			price:     500,
			percent:   0,
			wantPrice: 500,
			wantErr:   nil,
		},
		{
			name:      "hundred_percent_discount",
			price:     500,
			percent:   100,
			wantPrice: 0,
			wantErr:   nil,
		},
		{
			name:      "zero_price",
			price:     0,
			percent:   50,
			wantPrice: 0,
			wantErr:   nil,
		},
		{
			name:      "negative_price_returns_error",
			price:     -100,
			percent:   10,
			wantPrice: 0,
			wantErr:   ErrNegativePrice,
		},
		{
			name:      "percent_below_zero_returns_error",
			price:     1000,
			percent:   -5,
			wantPrice: 0,
			wantErr:   ErrInvalidPercent,
		},
		{
			name:      "percent_above_hundred_returns_error",
			price:     1000,
			percent:   150,
			wantPrice: 0,
			wantErr:   ErrInvalidPercent,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Apply(tt.price, tt.percent)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, 0, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPrice, got)
		})
	}
}

func TestFinalPrice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		items     []Item
		wantTotal int
		wantErr   error
	}{
		{
			name:      "empty_list",
			items:     []Item{},
			wantTotal: 0,
			wantErr:   nil,
		},
		{
			name:      "nil_list",
			items:     nil,
			wantTotal: 0,
			wantErr:   nil,
		},
		{
			name: "multiple_valid_items",
			items: []Item{
				{Price: 1000, DiscountPercent: 10}, // 900
				{Price: 500, DiscountPercent: 0},   // 500
				{Price: 200, DiscountPercent: 50},  // 100
				{Price: 300, DiscountPercent: 100}, // 0
			},
			wantTotal: 1500,
			wantErr:   nil,
		},
		{
			name: "item_with_invalid_price",
			items: []Item{
				{Price: 1000, DiscountPercent: 10},
				{Price: -50, DiscountPercent: 20},
			},
			wantTotal: 0,
			wantErr:   ErrNegativePrice,
		},
		{
			name: "item_with_invalid_discount_below_zero",
			items: []Item{
				{Price: 1000, DiscountPercent: 10},
				{Price: 500, DiscountPercent: -10},
			},
			wantTotal: 0,
			wantErr:   ErrInvalidPercent,
		},
		{
			name: "item_with_invalid_discount_above_hundred",
			items: []Item{
				{Price: 1000, DiscountPercent: 10},
				{Price: 500, DiscountPercent: 120},
			},
			wantTotal: 0,
			wantErr:   ErrInvalidPercent,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := FinalPrice(tt.items)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, 0, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantTotal, got)
		})
	}
}
