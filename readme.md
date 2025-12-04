

# Features

  * Basic functionality
    * [x] Slicing series
    * [x] Map and filter
    * [ ] Integration
    * [ ] Differentiation
    * [x] groupBy
    * [ ] Rolling window
    * [ ] Resampling
    * [x] Join 
    * [x] Merge

  * Calculate statistics
    * [x] Min, max
    * [x] Mean, variance and standard deviation
    * [ ] Covariance and correlation
    * [ ] Normalization

  * Data cleaning & missing data
    * [ ] Handling missing timestamps (gaps) in time axis
    * [ ] Forward-fill / back-fill of missing values
    * [ ] Interpolation (e.g. linear / step)
    * [ ] Simple outlier detection and clipping/removal

  * Time & indexing utilities
    * [ ] Reindexing series to a given time grid
    * [ ] Aggregation by periods (daily / weekly / monthly / custom)
    * [ ] Business calendar support (business days vs weekends)
    * [ ] Timezone-aware operations (convert, normalize to UTC)

  * Transformations & filters
    * [ ] Smoothing (moving average, exponential moving average)
    * [ ] Log transform / power transforms (e.g. Box–Cox)
    * [ ] Detrending (remove linear trend)
    * [ ] Deseasonalization (remove seasonal component)

  * Decomposition & spectral analysis
    * [ ] Time series decomposition: trend + seasonality + residual
    * [ ] FFT / power spectrum computation

  * Advanced statistics & features
    * [ ] Rolling / expanding statistics (rolling mean/var/min/max)
    * [ ] Exponentially weighted statistics (EWMA, EWVAR)
    * [ ] Feature generation for ML (lags, rolling features, calendar features)

  * Metrics
    * [x] MSE and RMSE between 2 series
    * [ ] MAE between 2 series
    * [ ] MAPE between 2 series
    * [ ] MAD between 2 series

  * ARIMA
    * [ ] Check if series is stationary
    * [ ] AR(p) – Autoregressive
    * [ ] I(d) – Integrate
    * [ ] MA(q) – Moving average

  * Forecasting (beyond ARIMA)
    * [ ] Naive forecast (last value, seasonal naive)
    * [ ] Simple / double / triple exponential smoothing (Holt–Winters)
    * [ ] Time-series cross-validation (walk-forward validation)

  * IO & interoperability
    * [ ] Read data to/from CSV string
    * [ ] Read/write CSV files
    * [ ] Read/write JSON (e.g. NDJSON)
    * [ ] Streaming read/write from io.Reader / io.Writer

  * Generators
    * [x] Constant series
    * [ ] Random noise
    * [x] Random walk
    * [ ] Periodic pattern

  * Anomaly detection
    * [ ] Z-score / robust z-score based detection
    * [ ] Threshold-based detection on residuals (e.g. from ARIMA or smoothing)
    * [ ] Simple rule-based anomaly flags (spikes, drops, flat-lines)

  * Advanced functionality
    * [ ] Finding sessions (periods of activity)
    * [ ] Labeling events/windows (e.g. storms, campaigns, outages)
