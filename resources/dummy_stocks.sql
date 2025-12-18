INSERT INTO stocks (stock_symbol, price) VALUES
    ('AAPL', 1523.45),
    ('GOOGL', 1688.20),
    ('MSFT', 1754.90),
    ('AMZN', 1812.75),
    ('TSLA', 1956.30),
    ('META', 1639.40),
    ('NFLX', 1877.60),
    ('NVDA', 1992.15),
    ('INTC', 1564.80),
    ('AMD', 1721.55)
ON CONFLICT (stock_symbol) DO NOTHING;