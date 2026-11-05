name: strtotime in a named default zone
description: >
  The same grammar read in Europe/Ljubljana from three base dates, daylight saving included. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
date_default_timezone_set('Europe/Ljubljana');
$bases = [strtotime('2024-03-15 10:30:45'), strtotime('2023-12-31 23:59:59'), strtotime('2024-02-29 00:00:00')];
$specs = [
 'noon tomorrow','midnight +1 day','tomorrow 08:15','today 18:00','next friday 5pm','last day of this month 23:59:59','first day of next year','last day of december','1 day ago 15:00',
 '2024-03-10','Mar 10','March 10th','10th March 2024','10-Mar-2024','10 Mar 2024 14:30','Mar-10-2024','2024-Mar-10','March 2024',
 'Thursday','thursday 3pm','this thursday','next thursday','last thursday','thu','sat','sunday next week','+2 weeks monday','monday +2 weeks','last monday of march 2024','first friday of 2024-05','third wednesday of next month',
 '+1 hour 15 minutes','-1 week','+5 days 3 hours','1 year 2 months ago','+1 month first day of','next month last day of','yesterday noon','yesterday midnight','tomorrow +3 hours',
 '12:30:45.5','12.30','9:5','9:05 am','12am','12 am','12pm','12:00 AM','0:00','noon yesterday',
 '2024-03-15 10:30:45 +1 day','2024-03-15T10:30:45.123456+01:00','2024-03-15 10:30:45 CET','2024-03-15 10:30:45 PST','2024-03-15 10:30:45 Asia/Tokyo','2024-03-15 10:30 GMT+2',
 '@0','@-86400','2000-13-01','2001-02-29',
 'sat next week','mon this week','tuesday last week', 'fortnight','next fortnight','last year december 25',
 'Dec 25','25 dec','25 december 2023 midnight','5/6/2024 7pm','5/6','2024.05.06','06.05.2024','6-5-2024',
 'eighth day','twelfth month','+0 day','0 days','now + 1 day','now - 1 day','tomorrow tomorrow',
];
foreach ($bases as $bi => $base) { foreach ($specs as $s) { $t = @strtotime($s, $base); echo "$bi ", str_pad("'$s'", 40), $t === false ? 'false' : date('Y-m-d H:i:s T', $t), "\n"; } }
---
0 'noon tomorrow'                         2024-03-16 00:00:00 CET
0 'midnight +1 day'                       2024-03-16 00:00:00 CET
0 'tomorrow 08:15'                        2024-03-16 08:15:00 CET
0 'today 18:00'                           2024-03-15 18:00:00 CET
0 'next friday 5pm'                       2024-03-22 17:00:00 CET
0 'last day of this month 23:59:59'       2024-03-31 23:59:59 CEST
0 'first day of next year'                2025-03-01 10:30:45 CET
0 'last day of december'                  2024-12-31 00:00:00 CET
0 '1 day ago 15:00'                       2024-03-14 15:00:00 CET
0 '2024-03-10'                            2024-03-10 00:00:00 CET
0 'Mar 10'                                2024-03-10 00:00:00 CET
0 'March 10th'                            2024-03-10 00:00:00 CET
0 '10th March 2024'                       2024-03-10 00:00:00 CET
0 '10-Mar-2024'                           2024-03-10 00:00:00 CET
0 '10 Mar 2024 14:30'                     2024-03-10 14:30:00 CET
0 'Mar-10-2024'                           2024-03-10 00:00:00 CET
0 '2024-Mar-10'                           2024-03-10 00:00:00 CET
0 'March 2024'                            2024-03-01 00:00:00 CET
0 'Thursday'                              2024-03-21 00:00:00 CET
0 'thursday 3pm'                          2024-03-21 15:00:00 CET
0 'this thursday'                         2024-03-21 00:00:00 CET
0 'next thursday'                         2024-03-21 00:00:00 CET
0 'last thursday'                         2024-03-14 00:00:00 CET
0 'thu'                                   2024-03-21 00:00:00 CET
0 'sat'                                   2024-03-16 00:00:00 CET
0 'sunday next week'                      2024-03-24 00:00:00 CET
0 '+2 weeks monday'                       2024-04-01 00:00:00 CEST
0 'monday +2 weeks'                       2024-04-01 00:00:00 CEST
0 'last monday of march 2024'             2024-03-25 00:00:00 CET
0 'first friday of 2024-05'               2024-05-03 00:00:00 CEST
0 'third wednesday of next month'         2024-04-17 00:00:00 CEST
0 '+1 hour 15 minutes'                    2024-03-15 11:45:45 CET
0 '-1 week'                               2024-03-08 10:30:45 CET
0 '+5 days 3 hours'                       2024-03-20 13:30:45 CET
0 '1 year 2 months ago'                   2023-01-15 10:30:45 CET
0 '+1 month first day of'                 2024-04-01 10:30:45 CEST
0 'next month last day of'                2024-04-30 10:30:45 CEST
0 'yesterday noon'                        2024-03-14 12:00:00 CET
0 'yesterday midnight'                    2024-03-14 00:00:00 CET
0 'tomorrow +3 hours'                     2024-03-16 03:00:00 CET
0 '12:30:45.5'                            2024-03-15 12:30:45 CET
0 '12.30'                                 2024-03-15 12:30:00 CET
0 '9:5'                                   2024-03-15 09:05:00 CET
0 '9:05 am'                               2024-03-15 09:05:00 CET
0 '12am'                                  2024-03-15 00:00:00 CET
0 '12 am'                                 2024-03-15 00:00:00 CET
0 '12pm'                                  2024-03-15 12:00:00 CET
0 '12:00 AM'                              2024-03-15 00:00:00 CET
0 '0:00'                                  2024-03-15 00:00:00 CET
0 'noon yesterday'                        2024-03-14 00:00:00 CET
0 '2024-03-15 10:30:45 +1 day'            2024-03-16 10:30:45 CET
0 '2024-03-15T10:30:45.123456+01:00'      2024-03-15 10:30:45 CET
0 '2024-03-15 10:30:45 CET'               2024-03-15 10:30:45 CET
0 '2024-03-15 10:30:45 PST'               2024-03-15 19:30:45 CET
0 '2024-03-15 10:30:45 Asia/Tokyo'        2024-03-15 02:30:45 CET
0 '2024-03-15 10:30 GMT+2'                2024-03-15 09:30:00 CET
0 '@0'                                    1970-01-01 01:00:00 CET
0 '@-86400'                               1969-12-31 01:00:00 CET
0 '2000-13-01'                            false
0 '2001-02-29'                            2001-03-01 00:00:00 CET
0 'sat next week'                         2024-03-23 00:00:00 CET
0 'mon this week'                         2024-03-11 00:00:00 CET
0 'tuesday last week'                     2024-03-05 00:00:00 CET
0 'fortnight'                             false
0 'next fortnight'                        2024-03-29 10:30:45 CET
0 'last year december 25'                 2023-12-25 00:00:00 CET
0 'Dec 25'                                2024-12-25 00:00:00 CET
0 '25 dec'                                2024-12-25 00:00:00 CET
0 '25 december 2023 midnight'             2023-12-25 00:00:00 CET
0 '5/6/2024 7pm'                          2024-05-06 19:00:00 CEST
0 '5/6'                                   2024-05-06 00:00:00 CEST
0 '2024.05.06'                            false
0 '06.05.2024'                            2024-05-06 00:00:00 CEST
0 '6-5-2024'                              2024-05-06 00:00:00 CEST
0 'eighth day'                            2024-03-23 10:30:45 CET
0 'twelfth month'                         2025-03-15 10:30:45 CET
0 '+0 day'                                2024-03-15 10:30:45 CET
0 '0 days'                                2024-03-15 10:30:45 CET
0 'now + 1 day'                           2024-03-16 10:30:45 CET
0 'now - 1 day'                           2024-03-14 10:30:45 CET
0 'tomorrow tomorrow'                     2024-03-16 00:00:00 CET
1 'noon tomorrow'                         2024-01-01 00:00:00 CET
1 'midnight +1 day'                       2024-01-01 00:00:00 CET
1 'tomorrow 08:15'                        2024-01-01 08:15:00 CET
1 'today 18:00'                           2023-12-31 18:00:00 CET
1 'next friday 5pm'                       2024-01-05 17:00:00 CET
1 'last day of this month 23:59:59'       2023-12-31 23:59:59 CET
1 'first day of next year'                2024-12-01 23:59:59 CET
1 'last day of december'                  2023-12-31 00:00:00 CET
1 '1 day ago 15:00'                       2023-12-30 15:00:00 CET
1 '2024-03-10'                            2024-03-10 00:00:00 CET
1 'Mar 10'                                2023-03-10 00:00:00 CET
1 'March 10th'                            2023-03-10 00:00:00 CET
1 '10th March 2024'                       2024-03-10 00:00:00 CET
1 '10-Mar-2024'                           2024-03-10 00:00:00 CET
1 '10 Mar 2024 14:30'                     2024-03-10 14:30:00 CET
1 'Mar-10-2024'                           2024-03-10 00:00:00 CET
1 '2024-Mar-10'                           2024-03-10 00:00:00 CET
1 'March 2024'                            2024-03-01 00:00:00 CET
1 'Thursday'                              2024-01-04 00:00:00 CET
1 'thursday 3pm'                          2024-01-04 15:00:00 CET
1 'this thursday'                         2024-01-04 00:00:00 CET
1 'next thursday'                         2024-01-04 00:00:00 CET
1 'last thursday'                         2023-12-28 00:00:00 CET
1 'thu'                                   2024-01-04 00:00:00 CET
1 'sat'                                   2024-01-06 00:00:00 CET
1 'sunday next week'                      2024-01-07 00:00:00 CET
1 '+2 weeks monday'                       2024-01-15 00:00:00 CET
1 'monday +2 weeks'                       2024-01-15 00:00:00 CET
1 'last monday of march 2024'             2024-03-25 00:00:00 CET
1 'first friday of 2024-05'               2024-05-03 00:00:00 CEST
1 'third wednesday of next month'         2024-01-17 00:00:00 CET
1 '+1 hour 15 minutes'                    2024-01-01 01:14:59 CET
1 '-1 week'                               2023-12-24 23:59:59 CET
1 '+5 days 3 hours'                       2024-01-06 02:59:59 CET
1 '1 year 2 months ago'                   2022-10-31 23:59:59 CET
1 '+1 month first day of'                 2024-01-01 23:59:59 CET
1 'next month last day of'                2024-01-31 23:59:59 CET
1 'yesterday noon'                        2023-12-30 12:00:00 CET
1 'yesterday midnight'                    2023-12-30 00:00:00 CET
1 'tomorrow +3 hours'                     2024-01-01 03:00:00 CET
1 '12:30:45.5'                            2023-12-31 12:30:45 CET
1 '12.30'                                 2023-12-31 12:30:00 CET
1 '9:5'                                   2023-12-31 09:05:00 CET
1 '9:05 am'                               2023-12-31 09:05:00 CET
1 '12am'                                  2023-12-31 00:00:00 CET
1 '12 am'                                 2023-12-31 00:00:00 CET
1 '12pm'                                  2023-12-31 12:00:00 CET
1 '12:00 AM'                              2023-12-31 00:00:00 CET
1 '0:00'                                  2023-12-31 00:00:00 CET
1 'noon yesterday'                        2023-12-30 00:00:00 CET
1 '2024-03-15 10:30:45 +1 day'            2024-03-16 10:30:45 CET
1 '2024-03-15T10:30:45.123456+01:00'      2024-03-15 10:30:45 CET
1 '2024-03-15 10:30:45 CET'               2024-03-15 10:30:45 CET
1 '2024-03-15 10:30:45 PST'               2024-03-15 19:30:45 CET
1 '2024-03-15 10:30:45 Asia/Tokyo'        2024-03-15 02:30:45 CET
1 '2024-03-15 10:30 GMT+2'                2024-03-15 09:30:00 CET
1 '@0'                                    1970-01-01 01:00:00 CET
1 '@-86400'                               1969-12-31 01:00:00 CET
1 '2000-13-01'                            false
1 '2001-02-29'                            2001-03-01 00:00:00 CET
1 'sat next week'                         2024-01-06 00:00:00 CET
1 'mon this week'                         2023-12-25 00:00:00 CET
1 'tuesday last week'                     2023-12-19 00:00:00 CET
1 'fortnight'                             false
1 'next fortnight'                        2024-01-14 23:59:59 CET
1 'last year december 25'                 2022-12-25 00:00:00 CET
1 'Dec 25'                                2023-12-25 00:00:00 CET
1 '25 dec'                                2023-12-25 00:00:00 CET
1 '25 december 2023 midnight'             2023-12-25 00:00:00 CET
1 '5/6/2024 7pm'                          2024-05-06 19:00:00 CEST
1 '5/6'                                   2023-05-06 00:00:00 CEST
1 '2024.05.06'                            false
1 '06.05.2024'                            2024-05-06 00:00:00 CEST
1 '6-5-2024'                              2024-05-06 00:00:00 CEST
1 'eighth day'                            2024-01-08 23:59:59 CET
1 'twelfth month'                         2024-12-31 23:59:59 CET
1 '+0 day'                                2023-12-31 23:59:59 CET
1 '0 days'                                2023-12-31 23:59:59 CET
1 'now + 1 day'                           2024-01-01 23:59:59 CET
1 'now - 1 day'                           2023-12-30 23:59:59 CET
1 'tomorrow tomorrow'                     2024-01-01 00:00:00 CET
2 'noon tomorrow'                         2024-03-01 00:00:00 CET
2 'midnight +1 day'                       2024-03-01 00:00:00 CET
2 'tomorrow 08:15'                        2024-03-01 08:15:00 CET
2 'today 18:00'                           2024-02-29 18:00:00 CET
2 'next friday 5pm'                       2024-03-01 17:00:00 CET
2 'last day of this month 23:59:59'       2024-02-29 23:59:59 CET
2 'first day of next year'                2025-02-01 00:00:00 CET
2 'last day of december'                  2024-12-31 00:00:00 CET
2 '1 day ago 15:00'                       2024-02-28 15:00:00 CET
2 '2024-03-10'                            2024-03-10 00:00:00 CET
2 'Mar 10'                                2024-03-10 00:00:00 CET
2 'March 10th'                            2024-03-10 00:00:00 CET
2 '10th March 2024'                       2024-03-10 00:00:00 CET
2 '10-Mar-2024'                           2024-03-10 00:00:00 CET
2 '10 Mar 2024 14:30'                     2024-03-10 14:30:00 CET
2 'Mar-10-2024'                           2024-03-10 00:00:00 CET
2 '2024-Mar-10'                           2024-03-10 00:00:00 CET
2 'March 2024'                            2024-03-01 00:00:00 CET
2 'Thursday'                              2024-02-29 00:00:00 CET
2 'thursday 3pm'                          2024-02-29 15:00:00 CET
2 'this thursday'                         2024-02-29 00:00:00 CET
2 'next thursday'                         2024-03-07 00:00:00 CET
2 'last thursday'                         2024-02-22 00:00:00 CET
2 'thu'                                   2024-02-29 00:00:00 CET
2 'sat'                                   2024-03-02 00:00:00 CET
2 'sunday next week'                      2024-03-10 00:00:00 CET
2 '+2 weeks monday'                       2024-03-18 00:00:00 CET
2 'monday +2 weeks'                       2024-03-18 00:00:00 CET
2 'last monday of march 2024'             2024-03-25 00:00:00 CET
2 'first friday of 2024-05'               2024-05-03 00:00:00 CEST
2 'third wednesday of next month'         2024-03-20 00:00:00 CET
2 '+1 hour 15 minutes'                    2024-02-29 01:15:00 CET
2 '-1 week'                               2024-02-22 00:00:00 CET
2 '+5 days 3 hours'                       2024-03-05 03:00:00 CET
2 '1 year 2 months ago'                   2022-12-29 00:00:00 CET
2 '+1 month first day of'                 2024-03-01 00:00:00 CET
2 'next month last day of'                2024-03-31 00:00:00 CET
2 'yesterday noon'                        2024-02-28 12:00:00 CET
2 'yesterday midnight'                    2024-02-28 00:00:00 CET
2 'tomorrow +3 hours'                     2024-03-01 03:00:00 CET
2 '12:30:45.5'                            2024-02-29 12:30:45 CET
2 '12.30'                                 2024-02-29 12:30:00 CET
2 '9:5'                                   2024-02-29 09:05:00 CET
2 '9:05 am'                               2024-02-29 09:05:00 CET
2 '12am'                                  2024-02-29 00:00:00 CET
2 '12 am'                                 2024-02-29 00:00:00 CET
2 '12pm'                                  2024-02-29 12:00:00 CET
2 '12:00 AM'                              2024-02-29 00:00:00 CET
2 '0:00'                                  2024-02-29 00:00:00 CET
2 'noon yesterday'                        2024-02-28 00:00:00 CET
2 '2024-03-15 10:30:45 +1 day'            2024-03-16 10:30:45 CET
2 '2024-03-15T10:30:45.123456+01:00'      2024-03-15 10:30:45 CET
2 '2024-03-15 10:30:45 CET'               2024-03-15 10:30:45 CET
2 '2024-03-15 10:30:45 PST'               2024-03-15 19:30:45 CET
2 '2024-03-15 10:30:45 Asia/Tokyo'        2024-03-15 02:30:45 CET
2 '2024-03-15 10:30 GMT+2'                2024-03-15 09:30:00 CET
2 '@0'                                    1970-01-01 01:00:00 CET
2 '@-86400'                               1969-12-31 01:00:00 CET
2 '2000-13-01'                            false
2 '2001-02-29'                            2001-03-01 00:00:00 CET
2 'sat next week'                         2024-03-09 00:00:00 CET
2 'mon this week'                         2024-02-26 00:00:00 CET
2 'tuesday last week'                     2024-02-20 00:00:00 CET
2 'fortnight'                             false
2 'next fortnight'                        2024-03-14 00:00:00 CET
2 'last year december 25'                 2023-12-25 00:00:00 CET
2 'Dec 25'                                2024-12-25 00:00:00 CET
2 '25 dec'                                2024-12-25 00:00:00 CET
2 '25 december 2023 midnight'             2023-12-25 00:00:00 CET
2 '5/6/2024 7pm'                          2024-05-06 19:00:00 CEST
2 '5/6'                                   2024-05-06 00:00:00 CEST
2 '2024.05.06'                            false
2 '06.05.2024'                            2024-05-06 00:00:00 CEST
2 '6-5-2024'                              2024-05-06 00:00:00 CEST
2 'eighth day'                            2024-03-08 00:00:00 CET
2 'twelfth month'                         2025-03-01 00:00:00 CET
2 '+0 day'                                2024-02-29 00:00:00 CET
2 '0 days'                                2024-02-29 00:00:00 CET
2 'now + 1 day'                           2024-03-01 00:00:00 CET
2 'now - 1 day'                           2024-02-28 00:00:00 CET
2 'tomorrow tomorrow'                     2024-03-01 00:00:00 CET
