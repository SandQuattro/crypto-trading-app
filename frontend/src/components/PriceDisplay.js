import React from 'react';

const PriceDisplay = ({ lastPrice, priceChange, symbol }) => {
  const isPositive = priceChange >= 0;
  const formattedPrice = lastPrice.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  });
  
  const formattedChange = priceChange.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
    signDisplay: 'always'
  });
  
  const changePercentage = ((priceChange / (lastPrice - priceChange)) * 100).toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
    signDisplay: 'always'
  });

  return (
    <div className="price-info">
      <div className="price">{symbol}: ${formattedPrice}</div>
      <div className={`price-change ${isPositive ? 'positive' : 'negative'}`}>
        {formattedChange} ({changePercentage}%)
      </div>
    </div>
  );
};

export default PriceDisplay;
