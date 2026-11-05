name: strtotime reads php's date grammar
description: >
  ISO, American and European dates, month names, clock times with a meridian, zones by name and offset, relative phrases, weekdays, first and last day of, and nth weekday of, against a fixed base. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
date_default_timezone_set('UTC');
$base = strtotime('2024-03-15 10:30:45');  // Friday
$specs = [
 'now','today','midnight','noon','tomorrow','yesterday','tomorrow noon','yesterday 14:00',
 '2000-01-02','2000-01-02 03:04:05','2000-01-02T03:04:05Z','2000-01-02T03:04:05+02:00','2000-01-02 03:04:05.678','2000/01/02','20000102','20000102T030405','2000-01',
 '1/2/2000','01/02/2000','1/2/00','1/2','1-2-2000','01-02-2000','1.2.2000','1.2.00','02.01.2000',
 '1 January 2000','January 1, 2000','Jan 1 2000','1st Jan 2000','Jan 1st, 2000','1 Jan','Jan 2000','January','Sat, 15 Aug 2015 12:00:00 +0000','Sat 15 Aug 2015','15 August 2015 3pm','2015-08-15 3:30 PM',
 '3pm','3:30pm','3:30 PM','15:52','15:52:01','3 pm','00:00','24:00',
 '+1 day','-1 day','+2 weeks','1 month ago','+1 year','+1 week 2 days','2 days ago','+90 minutes','+3 hours','-30 seconds','1 fortnight ago','+1 month -1 day',
 'next monday','last friday','monday','friday','next week','last week','next month','last month','next year','last year','this friday','monday next week','saturday this week','friday last week',
 'first day of next month','last day of next month','first day of this month','last day of february','last day of february 2020','first day of january 2025','first monday of next month','last friday of this month','second tuesday of april 2024',
 '@86400','@1700000000','@1700000000.5',
 'next month noon','2024-02-29 +1 year','2024-01-31 +1 month','first day of last month midnight',
 'Monday 10:00','tuesday 9am','12/25/2024 8:00','2024-06-01 12:00 UTC','2024-06-01 12:00 EST','2024-06-01 12:00 America/New_York','2024-06-01 12:00 +0530','10 September 2000','sept 3','dec 25 2024 noon',
 '+3 weekdays','garbage','', '1 hour 30 minutes ago', 'third friday of january 2025','eighth day','+1 week ago',
];
foreach ($specs as $s) { $t = strtotime($s, $base); echo str_pad("'$s'", 44), $t === false ? 'false' : date('Y-m-d H:i:s', $t), "\n"; }
---
'now'                                       2024-03-15 10:30:45
'today'                                     2024-03-15 00:00:00
'midnight'                                  2024-03-15 00:00:00
'noon'                                      2024-03-15 12:00:00
'tomorrow'                                  2024-03-16 00:00:00
'yesterday'                                 2024-03-14 00:00:00
'tomorrow noon'                             2024-03-16 12:00:00
'yesterday 14:00'                           2024-03-14 14:00:00
'2000-01-02'                                2000-01-02 00:00:00
'2000-01-02 03:04:05'                       2000-01-02 03:04:05
'2000-01-02T03:04:05Z'                      2000-01-02 03:04:05
'2000-01-02T03:04:05+02:00'                 2000-01-02 01:04:05
'2000-01-02 03:04:05.678'                   2000-01-02 03:04:05
'2000/01/02'                                2000-01-02 00:00:00
'20000102'                                  2000-01-02 00:00:00
'20000102T030405'                           2000-01-02 03:04:05
'2000-01'                                   2000-01-01 00:00:00
'1/2/2000'                                  2000-01-02 00:00:00
'01/02/2000'                                2000-01-02 00:00:00
'1/2/00'                                    2000-01-02 00:00:00
'1/2'                                       2024-01-02 00:00:00
'1-2-2000'                                  2000-02-01 00:00:00
'01-02-2000'                                2000-02-01 00:00:00
'1.2.2000'                                  2000-02-01 00:00:00
'1.2.00'                                    2024-03-15 01:02:00
'02.01.2000'                                2000-01-02 00:00:00
'1 January 2000'                            2000-01-01 00:00:00
'January 1, 2000'                           2000-01-01 00:00:00
'Jan 1 2000'                                2000-01-01 00:00:00
'1st Jan 2000'                              2000-01-01 00:00:00
'Jan 1st, 2000'                             2000-01-01 00:00:00
'1 Jan'                                     2024-01-01 00:00:00
'Jan 2000'                                  2000-01-01 00:00:00
'January'                                   2024-01-15 00:00:00
'Sat, 15 Aug 2015 12:00:00 +0000'           2015-08-15 12:00:00
'Sat 15 Aug 2015'                           2015-08-15 00:00:00
'15 August 2015 3pm'                        2015-08-15 15:00:00
'2015-08-15 3:30 PM'                        2015-08-15 15:30:00
'3pm'                                       2024-03-15 15:00:00
'3:30pm'                                    2024-03-15 15:30:00
'3:30 PM'                                   2024-03-15 15:30:00
'15:52'                                     2024-03-15 15:52:00
'15:52:01'                                  2024-03-15 15:52:01
'3 pm'                                      2024-03-15 15:00:00
'00:00'                                     2024-03-15 00:00:00
'24:00'                                     2024-03-16 00:00:00
'+1 day'                                    2024-03-16 10:30:45
'-1 day'                                    2024-03-14 10:30:45
'+2 weeks'                                  2024-03-29 10:30:45
'1 month ago'                               2024-02-15 10:30:45
'+1 year'                                   2025-03-15 10:30:45
'+1 week 2 days'                            2024-03-24 10:30:45
'2 days ago'                                2024-03-13 10:30:45
'+90 minutes'                               2024-03-15 12:00:45
'+3 hours'                                  2024-03-15 13:30:45
'-30 seconds'                               2024-03-15 10:30:15
'1 fortnight ago'                           2024-03-01 10:30:45
'+1 month -1 day'                           2024-04-14 10:30:45
'next monday'                               2024-03-18 00:00:00
'last friday'                               2024-03-08 00:00:00
'monday'                                    2024-03-18 00:00:00
'friday'                                    2024-03-15 00:00:00
'next week'                                 2024-03-18 10:30:45
'last week'                                 2024-03-04 10:30:45
'next month'                                2024-04-15 10:30:45
'last month'                                2024-02-15 10:30:45
'next year'                                 2025-03-15 10:30:45
'last year'                                 2023-03-15 10:30:45
'this friday'                               2024-03-15 00:00:00
'monday next week'                          2024-03-18 00:00:00
'saturday this week'                        2024-03-16 00:00:00
'friday last week'                          2024-03-08 00:00:00
'first day of next month'                   2024-04-01 10:30:45
'last day of next month'                    2024-04-30 10:30:45
'first day of this month'                   2024-03-01 10:30:45
'last day of february'                      2024-02-29 00:00:00
'last day of february 2020'                 2020-02-29 00:00:00
'first day of january 2025'                 2025-01-01 00:00:00
'first monday of next month'                2024-04-01 00:00:00
'last friday of this month'                 2024-03-29 00:00:00
'second tuesday of april 2024'              2024-04-09 00:00:00
'@86400'                                    1970-01-02 00:00:00
'@1700000000'                               2023-11-14 22:13:20
'@1700000000.5'                             2023-11-14 22:13:20
'next month noon'                           2024-04-15 12:00:00
'2024-02-29 +1 year'                        2025-03-01 00:00:00
'2024-01-31 +1 month'                       2024-03-02 00:00:00
'first day of last month midnight'          2024-02-01 00:00:00
'Monday 10:00'                              2024-03-18 10:00:00
'tuesday 9am'                               2024-03-19 09:00:00
'12/25/2024 8:00'                           2024-12-25 08:00:00
'2024-06-01 12:00 UTC'                      2024-06-01 12:00:00
'2024-06-01 12:00 EST'                      2024-06-01 17:00:00
'2024-06-01 12:00 America/New_York'         2024-06-01 16:00:00
'2024-06-01 12:00 +0530'                    2024-06-01 06:30:00
'10 September 2000'                         2000-09-10 00:00:00
'sept 3'                                    2024-09-03 00:00:00
'dec 25 2024 noon'                          2024-12-25 12:00:00
'+3 weekdays'                               2024-03-20 10:30:45
'garbage'                                   false
''                                          false
'1 hour 30 minutes ago'                     2024-03-15 09:00:45
'third friday of january 2025'              2025-01-17 00:00:00
'eighth day'                                2024-03-23 10:30:45
'+1 week ago'                               2024-03-08 10:30:45
