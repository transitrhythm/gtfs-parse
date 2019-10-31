package main

import (
	"fmt"
	"github.com/patrickbr/gtfsparser"      //"github.com/geops/gtfsparser"
	"github.com/patrickbr/gtfsparser/gtfs" //"github.com/geops/gtfsparser/gtfs"
	"log"
	"os"
	"sort"
	"strconv"
	"time"
)

// StopSchedule -
type StopSchedule struct {
	Route     *gtfs.Route
	StopTimes []*StopTime
}

// StopTime -
type StopTime struct {
	Route         *gtfs.Route
	Trip          *gtfs.Trip
	Service       *gtfs.Service
	ArrivalTime   gtfs.Time
	DepartureTime gtfs.Time
}

func stringtoInt(input string) (result int) {
	result, err := strconv.Atoi(input)
	if err != nil {
		result = 0
	}
	return result
}

// Schedules -
type Schedules []*StopSchedule

// StopSchedules -
var StopSchedules Schedules

// Len -
func (s Schedules) Len() int { return len(s) }

// Swap -
func (s Schedules) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// ByRouteID -
type ByRouteID struct{ Schedules }

// Less -
func (s ByRouteID) Less(i, j int) bool {
	return stringtoInt(s.Schedules[i].Route.Id) < stringtoInt(s.Schedules[j].Route.Id)
}

// // ByServiceID -
// type ByServiceID struct{ Schedules }

// // Less -
// func (s ByServiceID) Less(i, j int) bool {
// 	return stringtoInt(s.Schedules[i].Service.Id) < stringtoInt(s.Schedules[j].Service.Id)
// }

// AddToStopSchedule -
func AddToStopSchedule(route *gtfs.Route, stopTime StopTime) {
	for _, v := range StopSchedules {
		if v.Route.Id == route.Id {
			v.StopTimes = append(v.StopTimes, &stopTime)
			return
		}
	}
	stopSchedule := StopSchedule{}
	stopSchedule.Route = route
	stopSchedule.StopTimes = append(stopSchedule.StopTimes, &stopTime)
	StopSchedules = append(StopSchedules, &stopSchedule)
}

// toSeconds -
func toSeconds(input gtfs.Time) int {
	return (3600 * int(input.Hour)) + (60 * int(input.Minute)) + int(input.Second)
}

// StopTimes -
type StopTimes []*StopTime

// Len -
func (s StopTimes) Len() int { return len(s) }

// Swap -
func (s StopTimes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// ByArrivalTime -
type ByArrivalTime struct{ StopTimes }

// Less -
func (s ByArrivalTime) Less(i, j int) bool {
	return toSeconds(s.StopTimes[i].ArrivalTime) < toSeconds(s.StopTimes[j].ArrivalTime)
}

// Servicetable -
type Servicetable struct {
	Service *gtfs.Service
	Trips   []*gtfs.Trip
}

// Timetable -
type Timetable struct {
	StopTimes [7][]*StopTime
}

// TimetableA -
type TimetableA struct {
	serviceWeek ServiceWeek
}

// StopTimetable -
var StopTimetable Timetable

// Servicetables -
var Servicetables []*Servicetable

func addTripToServicetable(inputtables []*Servicetable, trip *gtfs.Trip) (outputtables []*Servicetable) {
	for _, v := range inputtables {
		if v.Service.Id == trip.Service.Id {
			v.Trips = append(v.Trips, trip)
			return inputtables
		}
	}
	servicetable := Servicetable{}
	servicetable.Service = trip.Service
	servicetable.Trips = append(servicetable.Trips, trip)
	outputtables = append(inputtables, &servicetable)
	return outputtables
}

// func tripService(feed *gtfsparser.Feed, tripID string) (service *gtfs.Service) {
// 	for _, v := range feed.Trips {
// 		if tripID == v.Id {
// 			service = v.Service
// 			break
// 		}
// 	}
// 	return service
// }

// // WeekdayServices -
// func WeekdayServices(feed *gtfsparser.Feed, tripID string) (daymap [7]bool) {
// 	service := tripService(feed, tripID)
// 	return service.Daymap
// }

func emptyDaymap(service *gtfs.Service) bool {
	for i := 0; i < len(service.Daymap); i++ {
		if service.Daymap[i] {
			return false
		}
	}
	return true
}

// ServiceDay - each day can be comprised a series of Service specifications
type ServiceDay []*gtfs.Service

// ServiceWeek - each service week consists of 7 days of daily Service specifications
type ServiceWeek [7]ServiceDay

// ServicesWeek -
var ServicesWeek ServiceWeek

// ServiceDelete -
func ServiceDelete(input ServiceDay, serviceID string) (output ServiceDay) {
	output = input
	for i, v := range output {
		if serviceID == v.Id {
			// Remove the element at index i from output.
			copy(output[i:], output[i+1:])  // Shift output[i+1:] left one index.
			output[len(output)-1] = nil     // Erase last element (write zero value).
			output = output[:len(output)-1] // Truncate slice.
			break
		}
	}
	return output
}

func createTimetable(feed *gtfsparser.Feed, stopCode string, schedules Schedules, weekEnding gtfs.Date) (timetable Timetable) {
	timetable = Timetable{}
	// Services are allocated on a weekly basis by a daily roster default
	for _, v := range Servicetables {
		days := v.Service.Daymap
		for i := 0; i < len(days); i++ {
			if days[i] {
				ServicesWeek[i] = append(ServicesWeek[i], v.Service)
			}
		}
	}
	// These defaults are modified on an exception basis for the provided specific dates
	if len(ServiceExceptions) > 0 {
		for _, v2 := range ServiceExceptions {
			weekday := int(toTime(v2.Date).Weekday())
			if v2.ExType == Add {
				ServicesWeek[weekday] = append(ServicesWeek[weekday], v2.Service)
			}
			if v2.ExType == Delete {
				ServicesWeek[weekday] = ServiceDelete(ServicesWeek[weekday], v2.Service.Id)
			}
		}
	}
	// The timetables are generated for the specified StopCode on a weekly basis
	stopID := getStopID(feed, stopCode)
	for i := 0; i < len(ServicesWeek); i++ {
		log.Printf("\nWeekday[%d]:", i)
		output := ""
		for _, v := range Servicetables {
			for _, v3 := range ServicesWeek[i] {
				if v.Service.Id == v3.Id {
					for _, v1 := range v.Trips {
						duration := time.Duration(7-i) * 24 * time.Hour
						weekdate := toTime(weekEnding).Add(-duration)
						if weekdate.After(toTime(v1.Service.Start_date)) && weekdate.Before(toTime(v1.Service.End_date)) {
							for _, v2 := range v1.StopTimes {
								if v2.Stop.Id == stopID {
									stopTime := StopTime{}
									stopTime.Route = v1.Route
									stopTime.Trip = v1
									stopTime.Service = v.Service
									stopTime.ArrivalTime = v2.Arrival_time
									stopTime.DepartureTime = v2.Departure_time
									output += fmt.Sprintf("[%s]%s-%d", v.Service.Id, v1.Route.Short_name, v2.Sequence)
									timetable.StopTimes[i] = append(timetable.StopTimes[i], &stopTime)
								}
							}
						}
					}
				}
			}
		}
		log.Println(output)
	}
	return timetable
}

// func createTimetableA(feed *gtfsparser.Feed, stopID string, schedules Schedules) (timetable Timetable) {
// 	// servicetable := Servicetable{}
// 	timetable = Timetable{}
// 	for _, v := range StopSchedules {
// 		for _, v1 := range v.StopTimes {
// 			// for _, v2 := range feed.Services {
// 			days := v1.Trip.Service.Daymap
// 			for i := 0; i < len(days); i++ {
// 				if days[i] {
// 					ServicesWeek[i] = v1.Trip.Service
// 				}
// 			}
// 			// services := WeekdayServices(feed, v1.Trip.Id)
// 			if len(ServiceExceptions) > 0 {
// 				for _, v2 := range ServiceExceptions {
// 					weekday := int(toTime(*v2.Date).Weekday())
// 					if v2.ExType == Add {
// 						ServicesWeek[weekday] = v2.Service
// 					}
// 				}
// 			}
// 				for i := 0; i < len(days); i++ {
// 					if days[i] {
// 						// servicetable.Timetables.StopTimes[i]
// 						timetable.StopTimes[i] = append(timetable.StopTimes[i], v1)
// 					}
// 				}
// 			}
// 			}
// 		}
// 	}
// 	return timetable
// }

func check(e error, message string) {
	if e != nil {
		if message != "" {
			log.Println(message)
		}
		panic(e)
	}
}

// TimeType -
type TimeType int

const (
	hhmmss TimeType = iota
	hhmm
	mm
)

//Timestamp -
func Timestamp(a gtfs.Time, t TimeType) (timestamp string) {
	if t == hhmmss {
		timestamp = fmt.Sprintf("%02d:%02d:%02d", a.Hour, a.Minute, a.Second)
	} else if t == hhmm {
		timestamp = fmt.Sprintf("%02d:%02d", a.Hour, a.Minute)
	} else if t == mm {
		timestamp = fmt.Sprintf(":%02d", a.Minute)
	}
	return
}

// TextStyle -
type TextStyle struct {
	textFrame string
	textFont  string
	textColor string
	color     string
}

func firstWords(value string, count int) string {
	// Loop over all indexes in the string.
	for i := range value {
		// If we encounter a space, reduce the count.
		if value[i] == ' ' {
			count--
			// When no more words required, return a substring.
			if count == 0 {
				return value[0:i]
			}
		}
	}
	// Return the entire string.
	return value
}

// TimeItem -
func TimeItem(feed *gtfsparser.Feed, item []*StopTime, max, index int) (routeName string, scheduleTime string) {
	var timeType TimeType
	if index > 0 && index < max && item[index-1].ArrivalTime.Hour == item[index].ArrivalTime.Hour {
		timeType = mm
	} else {
		timeType = hhmm
	}

	if index < max {
		routeName = item[index].Route.Short_name
		if len(routeName) == 0 {
			routeName = item[index].Trip.Short_name
		}
		if len(routeName) == 0 {
			routeName = firstWords(item[index].Trip.Headsign, 1)
		}
		scheduleTime = Timestamp(item[index].ArrivalTime, timeType)
	} else {
		routeName = ""
		scheduleTime = ""
	}
	return routeName, scheduleTime
}

// TimeItemCSV -
func TimeItemCSV(feed *gtfsparser.Feed, item []*StopTime, max, index int) (text string) {
	routeID, scheduleTime := TimeItem(feed, item, max, index)
	text = routeID + "," + scheduleTime + ","
	return text
}

func sortTimetable(timetable Timetable) {
	for i := 0; i < len(timetable.StopTimes); i++ {
		sort.Sort(ByArrivalTime{timetable.StopTimes[i]})
	}
}

// DayOfWeek -
type DayOfWeek struct {
	name   string
	abbrev string
}

// DaysOfWeek -
type DaysOfWeek struct {
	lang       string
	daysOfWeek [7]DayOfWeek
}

var daysOfWeekPt = [7]DayOfWeek{
	DayOfWeek{"Dominga", "DOM"},
	DayOfWeek{"Segunda-feira", "SEG"},
	DayOfWeek{"Terça-feira", "TER"},
	DayOfWeek{"Quarta-feira", "QUA"},
	DayOfWeek{"Quinta-feira", "QUI"},
	DayOfWeek{"Sexta-feira", "SEX"},
	DayOfWeek{"Sábado", "SÁB"},
}

var daysOfWeekEs = [7]DayOfWeek{
	DayOfWeek{"Dominga", "DOM"},
	DayOfWeek{"Lunes", "LUN"},
	DayOfWeek{"Martes", "MAR"},
	DayOfWeek{"Miércoles", "MIÉ"},
	DayOfWeek{"Jueves", "JEU"},
	DayOfWeek{"Viernes", "VIE"},
	DayOfWeek{"Sábado", "SÁB"},
}

var daysOfWeekFr = [7]DayOfWeek{
	DayOfWeek{"Dimanche", "DIM"},
	DayOfWeek{"Lundi", "LUN"},
	DayOfWeek{"Mardi", "MAR"},
	DayOfWeek{"Mercredi", "MER"},
	DayOfWeek{"Jeudi", "JEU"},
	DayOfWeek{"Vendredi", "VEN"},
	DayOfWeek{"Samedi", "SAM"},
}

var daysOfWeekEn = [7]DayOfWeek{
	DayOfWeek{"Sunday", "SUN"},
	DayOfWeek{"Monday", "MON"},
	DayOfWeek{"Tuesday", "TUE"},
	DayOfWeek{"Wednesday", "WED"},
	DayOfWeek{"Thursday", "THU"},
	DayOfWeek{"Friday", "FRI"},
	DayOfWeek{"Saturday", "SAT"},
}

var daysOfWeekDef = [7]DayOfWeek{
	DayOfWeek{"1", "1"},
	DayOfWeek{"2", "2"},
	DayOfWeek{"3", "3"},
	DayOfWeek{"4", "4"},
	DayOfWeek{"5", "5"},
	DayOfWeek{"6", "6"},
	DayOfWeek{"7", "7"},
}

var daysOfWeekList = []DaysOfWeek{
	{"en", daysOfWeekEn},
	{"es", daysOfWeekEs},
	{"pt", daysOfWeekPt},
	{"fr", daysOfWeekFr},
	{"def", daysOfWeekDef},
}

func toDate(year int, month time.Month, day int) gtfs.Date {
	return gtfs.Date{int8(day), int8(month), int16(year)}
}

//Datestamp -
func Datestamp(a gtfs.Date) (datestamp string) {
	return fmt.Sprintf("%04d-%02d-%02d", a.Year, a.Month, a.Day)
}

func findDaysOfWeekAbbrev(feed *gtfsparser.Feed, agency *gtfs.Agency) (days [7]string) {
	for _, v1 := range daysOfWeekList {
		if v1.lang == agency.Lang.GetLangString() {
			for k2, v2 := range v1.daysOfWeek {
				days[k2] = v2.abbrev
			}
			break
		}
	}
	return days
}

// ScheduleField -
type ScheduleField struct {
	routeID     string
	arrivalTime string
}

// ScheduleRow -
type ScheduleRow struct {
	fields [8]ScheduleField
}

func printTimetableHTML(feed *gtfsparser.Feed, filename string, timetable Timetable, stopCode string) (row ScheduleRow) {
	tableLength := 0
	for i := 0; i < len(timetable.StopTimes); i++ {
		if len(timetable.StopTimes[i]) > tableLength {
			tableLength = len(timetable.StopTimes[i])
		}
	}
	for i := 0; i < tableLength; i++ {
		row = ScheduleRow{}
		for j := 0; j < len(row.fields); j++ {
			field := ScheduleField{}
			field.routeID, field.arrivalTime = TimeItem(feed, timetable.StopTimes[j], len(timetable.StopTimes[j]), i)
			row.fields[i] = field
		}
	}
	return row
}

func printTimetableCSV(feed *gtfsparser.Feed, filename string, timetable Timetable, stop *gtfs.Stop, weekEnding gtfs.Date) (err error) {
	for _, feedInfo := range feed.FeedInfos {
		title := feedInfo.Publisher_name + "\n" + "Version:" + feedInfo.Version + "\n"
		for _, agency := range feed.Agencies {
			d := findDaysOfWeekAbbrev(feed, agency)
			file, err := os.Create(agency.Name + "-" + filename)
			check(err, "File create error")
			defer file.Close()

			tableLength := 0
			for i := 0; i < len(timetable.StopTimes); i++ {
				if len(timetable.StopTimes[i]) > tableLength {
					tableLength = len(timetable.StopTimes[i])
				}
			}
			start, end := getFeedDateRange(feed, 0)
			header := fmt.Sprintf("%s\nTransit Schedule\nStop #%s - %s\nFrom: %s - To: %s\n%s\n#,%s,#,%s,#,%s,#,%s,#,%s,#,%s,#,%s\n", agency.Name, stop.Code, stop.Desc, Datestamp(start), Datestamp(end), WeekDateCSV(timetable, weekEnding), d[time.Monday], d[time.Tuesday], d[time.Wednesday], d[time.Thursday], d[time.Friday], d[time.Saturday], d[time.Sunday])
			file.WriteString(title + header)
			log.Printf(title + header)

			for index := 0; index < tableLength; index++ {
				var line string
				for weekday := time.Monday; weekday <= time.Saturday; weekday++ {
					line += TimeItemCSV(feed, timetable.StopTimes[weekday], len(timetable.StopTimes[weekday]), index)
				}
				line += TimeItemCSV(feed, timetable.StopTimes[time.Sunday], len(timetable.StopTimes[time.Sunday]), index)
				line += "\n"
				_, err = file.WriteString(line)
				log.Printf("%s", line)
				check(err, "File write error")
				file.Sync()
			}
		}
	}
	return err
}

func toDayOfYear(date gtfs.Date) int {
	time := time.Date(int(date.Year), time.Month(date.Month), int(date.Day), 0, 0, 0, 0, time.Local)
	return time.YearDay()
}

func getFeedDateRange(feed *gtfsparser.Feed, index int) (start, end gtfs.Date) {
	start = feed.FeedInfos[index].Start_date
	end = feed.FeedInfos[index].End_date
	return start, end
}

// func getServiceDateRange(feed *gtfsparser.Feed, serviceID string) (start, end gtfs.Date) {
// 	for _, service := range feed.Services {
// 		if service.Id == serviceID {
// 			start = service.Start_date
// 			end = service.End_date
// 		}
// 	}
// 	return start, end
// }

// // FeedServices -
// type FeedServices []*Feed.Services

// // Services -
// var Services FeedServices

// // Len -
// func (s FeedServices) Len() int { return len(s) }

// // Swap -
// func (s FeedServices) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// // ByServiceID -
// type ByServiceID struct{ FeedServices }

// // Less -
// func (s ByServiceID) Less(i, j int) bool {
// 	return stringtoInt(s.Schedules[i].RouteID) < stringtoInt(s.Schedules[j].RouteID)
// }

func nextMonday() (date gtfs.Date) {
	today := time.Now()
	dayOfWeek := today.Weekday()
	daysUntilMonday := time.Duration(8 - dayOfWeek)
	if dayOfWeek == time.Sunday {
		daysUntilMonday = 1
	}
	monday := today.Add(daysUntilMonday * 24 * time.Hour)
	year, month, day := monday.Date()
	return toDate(year, month, day)
}

func thisSunday() gtfs.Date {
	year, month, day := time.Now().Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	dayOfWeek := time.Now().Weekday()
	if dayOfWeek != time.Sunday {
		sunday := today.Add(time.Hour * time.Duration(7-dayOfWeek) * 24)
		year, month, day = sunday.Date()
	}
	return toDate(year, month, day)
}

func thisMonth() gtfs.Date {
	year, month, _ := time.Now().Date()
	return toDate(year, month, 1)
}

const (
	layoutISO = "2006-01-02"
	layoutUS  = "Monday 2-January-2006"
)

// WeekEnding -
func WeekEnding(weekEnding gtfs.Date) (text string) {
	isoFormat := Datestamp(weekEnding)
	t, _ := time.Parse(layoutISO, isoFormat)
	text = "Week Ending: " + t.Format(layoutUS)
	return text
}

func daysThisMonth() (days int) {
	daysInMonth := [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	year, month, _ := time.Now().Date()
	days = daysInMonth[month-1]
	isLeapYear := time.Date(year, time.December, 31, 0, 0, 0, 0, time.Local).YearDay() == 366
	if month == time.February && isLeapYear {
		days = 29
	}
	return days
}

func dayOfThisMonth(dayOfMonth int) (abbrev string) {
	return abbrev
}

// WeekDates -
func WeekDates(weekEnding gtfs.Date) (days [7]string) {
	daysInMonth := [12]int8{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	day := weekEnding.Day
	month := int(weekEnding.Month)
	for i := len(days) - 1; i >= 0; i-- {
		days[i] = strconv.Itoa(int(day))
		day--
		if day < 1 {
			if month == 1 {
				month = 12
			} else {
				month--
			}
			day = daysInMonth[month]
		}
	}
	return days
}

// WeekDatesCSV -
func WeekDatesCSV(weekEnding gtfs.Date) (weekDates string) {
	days := WeekDates(weekEnding)
	for _, day := range days {
		weekDates += ",,,," + day + ","
	}
	return weekDates
}

// WeekDateRibbonCSV -
func WeekDateRibbonCSV(timetable Timetable, weekEnding gtfs.Date) (csv string) {
	days := WeekDates(weekEnding)
	for i := time.Monday; i <= time.Saturday; i++ {
		if len(timetable.StopTimes[i]) != 0 {
			csv += (timetable.StopTimes[i][0].Service.Id + "," + days[i-1] + ",")
		}
	}
	csv += (timetable.StopTimes[time.Sunday][0].Service.Id + "," + days[6] + ",")
	return csv
}

// WeekDateCSV -
func WeekDateCSV(timetable Timetable, weekEnding gtfs.Date) (text string) {
	log.Printf("\nWeekDateCSV():")
	text = WeekEnding(weekEnding) + "\n" + WeekDateRibbonCSV(timetable, weekEnding)
	log.Printf("%s - end", text)
	return text
}

// Exception -
type Exception int

// TripDirection -
type TripDirection int

const (
	// Add - add service exception
	Add Exception = iota + 1
	// Delete - delete service exception
	Delete
	// Inbound -
	Inbound TripDirection = iota
	// Outbound -
	Outbound
)

// ExceptionType -
func ExceptionType(exception Exception) (value string) {
	switch exception {
	case 1:
		value = "Add"
	case 2:
		value = "Delete"
	}
	return value
}

// // ExceptionDate -
// type ExceptionDate map[gtfs.Date]int8

// // ServiceExceptions -
// type ServiceExceptions []*ExceptionDate

// // Len -
// func (s ServiceExceptions) Len() int { return len(s) }

// // Swap -
// func (s ServiceExceptions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// // ByExceptionDate -
// type ByExceptionDate struct{ ServiceExceptions }

// // Less -
// func (s ByExceptionDate) Less(i, j int) bool {
// 	a := s.ServiceExceptions[i]
// 	fmt.Println(a)
// 	return gtfs.GetTime(s.ServiceExceptions[i][gtfs.Date]) < gtfs.GetTime(s.ServiceExceptions[j][gtfs.Date])
// }

var (
	// OfficialStartOfDayTime -
	OfficialStartOfDayTime = gtfs.Time{4, 0, 0} // 4am
)

// Convert GTFS Date to Time @ 12:00:00 noon local time
func toTime(date gtfs.Date) time.Time {
	return time.Date(int(date.Year), time.Month(date.Month), int(date.Day), 12, 0, 0, 0, time.Local)
}

// Convert GTFS Date to Time @ OfficialStartOfDayTime local time
func toStartTime(date gtfs.Date) time.Time {
	return time.Date(int(date.Year), time.Month(date.Month), int(date.Day), int(OfficialStartOfDayTime.Hour), int(OfficialStartOfDayTime.Minute), int(OfficialStartOfDayTime.Second), 0, time.Local)
}

// thisWorkingWeek - returns true if the specified date & time in in range of the current Timetable Working Week from OfficialStartOfDayTime from (Monday to Monday)
func thisWorkingWeek(date gtfs.Date) (inRange bool) {
	specifiedTime := toTime(date)
	endOfWeek := toStartTime(nextMonday())
	startOfWeek := endOfWeek.Add(-time.Hour * 7 * 24)
	inRange = (specifiedTime.After(startOfWeek) && specifiedTime.Before(endOfWeek))
	return inRange
}

// ServiceException -
type ServiceException struct {
	ExType  Exception
	Service *gtfs.Service
	Date    gtfs.Date
}

// ServiceExceptions -
var ServiceExceptions []*ServiceException

func addToExceptionTable(service *gtfs.Service, date *gtfs.Date, exception Exception) {
	serviceException := ServiceException{}
	serviceException.ExType = exception
	serviceException.Service = service
	serviceException.Date = *date
	ServiceExceptions = append(ServiceExceptions, &serviceException)
}

func getStopID(feed *gtfsparser.Feed, stopCode string) string {
	for _, v := range feed.Stops {
		if stopCode == v.Code {
			return v.Id
		}
	}
	return ""
}

func testSuite() {
	if !thisWorkingWeek(toDate(2019, 10, 20)) {
		log.Println("Pass")
	}
	if !thisWorkingWeek(toDate(2019, 10, 28)) {
		log.Println("Pass")
	}
	if thisWorkingWeek(toDate(2019, 10, 21)) {
		log.Println("Pass")
	}
	if thisWorkingWeek(toDate(2019, 10, 27)) {
		log.Println("Pass")
	}
}

func createLogFile() *os.File {
	file, err := os.OpenFile("GTFS-Parse.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Log open failure")
	}
	log.SetOutput(file)
	log.Println("GTFS-Parse started")
	return file
}

func findStop(feed *gtfsparser.Feed, stopCode string) (stop *gtfs.Stop) {
	for _, stop = range feed.Stops {
		if stopCode == stop.Code {
			break
		}
	}
	return stop
}

func setupFeed(zipFile string) (feed *gtfsparser.Feed) {
	feed = gtfsparser.NewFeed()
	feed.Parse(zipFile)
	return feed
}

func processStops(feed *gtfsparser.Feed, StopSchedules Schedules, stopCode string) {
	stop := findStop(feed, stopCode)
	timetable := createTimetable(feed, stop.Code, StopSchedules, thisSunday())
	sortTimetable(timetable)
	printTimetableCSV(feed, "Timetable-"+stopCode+"-WE-"+Datestamp(thisSunday())+".csv", timetable, stop, thisSunday())
}

func processBlocks(feed *gtfsparser.Feed, Blocktables []*Blocktable, blockID string) {
	blockSchedule := createBlockSchedule(Blocktables, blockID, nextMonday())
	sortBlockSchedule(blockSchedule)
	printBlockWeekCSV(feed, "BlockWeek-"+blockID+"-WE-"+Datestamp(thisSunday())+".csv", &blockSchedule, blockID, thisSunday())
	blockCalendar := createBlockCalendar(Blocktables)
	sortBlockCalendar(blockCalendar)
	printBlockMonthCSV(feed, "BlockMonth-"+Datestamp(thisMonth())+".csv", &blockCalendar, thisMonth())
	BlockSchedules = append(BlockSchedules, &blockSchedule)
	deadheadSchedule := createDeadheadSchedule(Blocktables, nextMonday())
	printDeadheadWeekCSV(feed, "DeadheadWeek-"+Datestamp(thisSunday())+".csv", &deadheadSchedule, thisSunday())
}

func processAgency(feed *gtfsparser.Feed, stopCode string, blockID string) {
	for _, trip := range feed.Trips {
		sort.Sort(trip.StopTimes)
		Servicetables = addTripToServicetable(Servicetables, trip)
		Blocktables = addTripToBlocktable(Blocktables, trip)
	}

	// for _, v := range feed.Stops {
	// 	if stopCode == v.Code {
	// 		stopID := v.Id
	// 		// log.Printf("[%s] %s : %s (@ %f,%f)\n", v.Code, k, v.Name, v.Lat, v.Lon)
	// 		stopTime := StopTime{}
	// 		for _, trip := range feed.Trips {
	// 			for _, times := range trip.StopTimes {
	// 				if times.Stop.Id == stopID {
	// 					stopTime.Route = trip.Route
	// 					stopTime.Trip = trip
	// 					stopTime.Service = trip.Service
	// 					stopTime.ArrivalTime = times.Arrival_time
	// 					stopTime.DepartureTime = times.Departure_time
	// 					AddToStopSchedule(trip.Route, stopTime)
	// 					break
	// 				}
	// 			}
	// 		}
	// 	}
	// }
	// // Sort the Stop Schedule by Route ID and by Stop Arrival Time
	// sort.Sort(ByRouteID{StopSchedules})
	// for _, v := range StopSchedules {
	// 	// fmt.Println("Route:", v.Route.Id)
	// 	sort.Sort(ByArrivalTime{v.StopTimes})
	// 	// for i, time := range v.StopTimes {
	// 	// 	fmt.Println("[", i, "]", time.Service.Id, time.ArrivalTime)
	// 	// }
	// }

	for _, service := range feed.Services {
		log.Println("Service: ", service.Id, Datestamp(service.Start_date), Datestamp(service.End_date))
		log.Println("First: ", Datestamp(service.GetFirstDefinedDate()), "Last: ", Datestamp(service.GetLastDefinedDate()))
		// sort.Sort(ByExceptionDate{service.Exceptions})
		for k2, v := range service.Exceptions {
			if thisWorkingWeek(k2) {
				addToExceptionTable(service, &k2, Exception(v))
			}
			log.Println("Exception:", service.Id, Datestamp(k2), ExceptionType(Exception(v)))
		}
	}
	processStops(feed, StopSchedules, stopCode)
	processBlocks(feed, Blocktables, blockID)
}

var currentFeed *gtfsparser.Feed

func setCurrentFeed(feed *gtfsparser.Feed) {
	currentFeed = feed
}

func GetCurrentFeed() *gtfsparser.Feed {
	return currentFeed
}

func main() {
	file := createLogFile()
	defer file.Close()

	testSuite()

	log.Println(len(os.Args), os.Args)
	if len(os.Args) != 4 {
		log.Printf("Usage : %d - %s <ZIPfile> <StopCode> <BlockID>\n", len(os.Args), os.Args)
		os.Exit(0)
	}

	zipFile := os.Args[1]
	// OfficialStartOfDayTime := os.Args[2]
	stopCode := os.Args[2]
	blockID := os.Args[3]
	feed := setupFeed(zipFile)
	setCurrentFeed(feed)
	log.Printf("Done, parsed %d agencies, %d stops, %d routes, %d trips, %d fare attributes\n\n",
		len(feed.Agencies), len(feed.Stops), len(feed.Routes), len(feed.Trips), len(feed.FareAttributes))

	for agency := 0; agency < len(feed.Agencies); agency++ {
		processAgency(feed, stopCode, blockID)
	}
	httpServer()
}
