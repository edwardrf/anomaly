Simple Periodic data anomaly detection implemented in golang
============================================================

This implementation of anomaly detection use FFT to try to determine if the data has a repeating pattern, if there is, try to determine its period with the FFT result. If a repeating period cannot be determined, the algorithm will use a simple 3 sigma rule, reporting datapoints 3 standard deviation away from its mean. If there is a pattern, all data points are chopped into segments of length of one period, and all data points at the exact same location within one period is checked with the 3 sigma rule.

Test data
---------
Test data is copied from the NAB project (https://github.com/numenta/NAB), which is AGPL licensed.
