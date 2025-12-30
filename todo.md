# TODO

Missing or planned functionality.

## Statistics
- Covariance and correlation
- Normalization helpers (min-max, scaling)

## Data cleaning and missing data
- Handling missing timestamps (gaps) in time axis
- Forward-fill / back-fill of missing values
- Simple outlier detection and clipping/removal

## Time and indexing utilities
- Reindexing series to a given time grid
- Aggregation by periods (daily / weekly / monthly / custom)
- Business calendar support (business days vs weekends)
- Timezone-aware operations (convert, normalize to UTC)

## Transformations and filters
- Exponential moving average (EMA) and other smoothing helpers
- Log transform / power transforms (e.g. Box-Cox)
- Detrending (remove linear trend)
- Deseasonalization (remove seasonal component)

## Decomposition and spectral analysis
- Time series decomposition: trend + seasonality + residual
- FFT / power spectrum computation

## Advanced statistics and features
- Rolling / expanding statistics helpers (rolling mean/var/min/max)
- Exponentially weighted statistics (EWMA, EWVAR)
- Feature generation for ML (lags, rolling features, calendar features)

## Metrics
- MAPE between 2 series

## ARIMA
- Check if series is stationary
- AR(p) - Autoregressive
- I(d) - Integrate
- MA(q) - Moving average

## Forecasting
- Seasonal naive forecast
- Simple / double / triple exponential smoothing (Holt-Winters)
- Time-series cross-validation (walk-forward validation)

## IO and interoperability
- Read/write JSON (e.g. NDJSON)

## Generators
- Random noise

## Anomaly detection
- Threshold-based detection on residuals (e.g. from ARIMA or smoothing)

## Advanced functionality
- Finding sessions (periods of activity)
- Labeling events/windows (e.g. storms, campaigns, outages)
