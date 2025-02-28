import React, { useState, useEffect } from 'react';
import TradingChart from './components/TradingChart';
import PairSelector from './components/PairSelector';
import PriceDisplay from './components/PriceDisplay';
import { fetchTradingPairs } from './services/api';

function App() {
  const [tradingPairs, setTradingPairs] = useState([]);
  const [selectedPair, setSelectedPair] = useState('');
  const [pairData, setPairData] = useState({
    lastPrice: 0,
    priceChange: 0
  });
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    const loadTradingPairs = async () => {
      try {
        const pairs = await fetchTradingPairs();
        setTradingPairs(pairs);
        
        if (pairs.length > 0) {
          const btcPair = pairs.find(p => p.symbol === 'BTCUSDT');
          const initialPair = btcPair || pairs[0];
          
          setPairData({
            lastPrice: initialPair.lastPrice,
            priceChange: initialPair.priceChange
          });
          
          setTimeout(() => {
            setSelectedPair(initialPair.symbol);
            setIsLoaded(true);
          }, 500);
        }
      } catch (error) {
        console.error('Error loading trading pairs:', error);
      }
    };

    loadTradingPairs();
  }, []);

  useEffect(() => {
    if (!selectedPair) return;

    const ws = new WebSocket(`ws://localhost:8080/ws/${selectedPair}`);
    
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      setPairData({
        lastPrice: data.lastPrice,
        priceChange: data.priceChange
      });
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    return () => {
      ws.close();
    };
  }, [selectedPair]);

  const handleSelectPair = (symbol) => {
    setSelectedPair(symbol);
    
    const pair = tradingPairs.find(p => p.symbol === symbol);
    if (pair) {
      setPairData({
        lastPrice: pair.lastPrice,
        priceChange: pair.priceChange
      });
    }
  };

  return (
    <div className="app-container">
      <header className="header">
        <h1>Crypto Trading</h1>
      </header>
      <div className="trading-container">
        <PairSelector 
          pairs={tradingPairs
            .map(p => p.symbol)
            .sort((a, b) => {
              // Fixed order of pairs
              const order = {
                'BTCUSDT': 1,
                'ETHUSDT': 2,
                'SOLUSDT': 3,
                'BNBUSDT': 4,
                'XRPUSDT': 5
              };
              return (order[a] || 999) - (order[b] || 999);
            })} 
          selectedPair={selectedPair} 
          onSelectPair={handleSelectPair} 
        />
        <PriceDisplay 
          symbol={selectedPair}
          lastPrice={pairData.lastPrice}
          priceChange={pairData.priceChange}
        />
        {isLoaded && selectedPair && (
          <div className="chart-container">
            <TradingChart 
              key={selectedPair} 
              symbol={selectedPair} 
            />
          </div>
        )}
      </div>
    </div>
  );
}

export default App;
