package main

import (
	"encoding/json"
	"fmt"
	"github.com/patrickbr/gtfsparser"      //"github.com/geops/gtfsparser"
	"github.com/patrickbr/gtfsparser/gtfs" //"github.com/geops/gtfsparser/gtfs"
	"log"
	"os"
	"sort"
	"strconv"
	"time"
)

// Blocktable -
type Blocktable struct {
	BlockID       string
	Servicetables []*Servicetable
}

// Blocktables -
var Blocktables []*Blocktable

// Block -
type Block struct {
	BlockID string
	StartAt gtfs.Time
	EndAt   gtfs.Time
	Trips   []*gtfs.Trip
}

// BlockDay -
type BlockDay struct {
	Date   gtfs.Date
	Blocks []*Block
}

// BlockCalendar -
type BlockCalendar struct {
	BlockDays []*BlockDay
}

// BlockSchedule -
type BlockSchedule struct {
	BlockDays []*BlockDay //	Blocks [7]*Block
}

// BlockSchedules -
var BlockSchedules []*BlockSchedule

// BlocktableDay -
type BlocktableDay Blocktable

// BlocktableWeek -
type BlocktableWeek [7]*BlocktableDay

// BlocktablesWeek -
var BlocktablesWeek BlocktableWeek

func addServicetableToBlocktable(blocktable *Blocktable, servicetable *Servicetable) (outputtable *Blocktable) {

	return outputtable
}

func addTripToBlocktable(blocktables []*Blocktable, trip *gtfs.Trip) (outputtables []*Blocktable) {
	if trip.Block_id == "51841" {
		log.Printf("addTripToBlocktable() - Trip ID: %s Block ID: %s Service ID: %s\n", trip.Id, trip.Block_id, trip.Service.Id)
	}
	for _, blocktable := range blocktables {
		if blocktable.BlockID == trip.Block_id {
			var servicetables []*Servicetable
			for _, servicetable := range blocktable.Servicetables {
				if servicetable.Service.Id == trip.Service.Id {
					servicetables = addTripToServicetable(blocktable.Servicetables, trip)
					blocktable.Servicetables = servicetables
					break
				}
			}
			if servicetables == nil {
				servicetable := Servicetable{}
				servicetable.Service = trip.Service
				servicetable.Trips = append(servicetable.Trips, trip)
				blocktable.Servicetables = append(blocktable.Servicetables, &servicetable)
				// addServicetableToBlocktable(blocktable, &servicetable)
			}
			return blocktables
		}
	}
	blocktable := Blocktable{}
	blocktable.BlockID = trip.Block_id
	blocktable.Servicetables = addTripToServicetable(blocktable.Servicetables, trip)
	outputtables = append(blocktables, &blocktable)
	return outputtables
}

// BlockItem -
func BlockItem(feed *gtfsparser.Feed, trips []*gtfs.Trip, max, index int) (routeName, tripID, serviceID, scheduleTime, directionID string) {
	var timeType TimeType
	if index > 0 && index < max && trips[index-1].StopTimes[0].Departure_time.Hour == trips[index].StopTimes[0].Departure_time.Hour {
		timeType = mm
	} else {
		timeType = hhmm
	}

	if index < max {
		routeName = trips[index].Route.Short_name
		if len(routeName) == 0 {
			routeName = trips[index].Short_name
		}
		if len(routeName) == 0 {
			routeName = firstWords(trips[index].Headsign, 1)
		}
		scheduleTime = Timestamp(trips[index].StopTimes[0].Departure_time, timeType)
		directionID = strconv.Itoa(int(trips[index].Direction_id))
		serviceID = trips[index].Service.Id
		tripID = trips[index].Id
	} else {
		tripID = ""
		routeName = ""
		serviceID = ""
		directionID = ""
		scheduleTime = ""
	}
	return routeName, tripID, serviceID, scheduleTime, directionID
}

// BlockItemCSV -
func BlockItemCSV(feed *gtfsparser.Feed, trips []*gtfs.Trip, max, index int) (text string) {
	routeName, tripID, serviceID, startTime, directionID := BlockItem(feed, trips, max, index)
	text = routeName + "," + tripID + "," + serviceID + "," + directionID + "," + startTime + ","
	return text
}

// BlockCalendarItem -
func BlockCalendarItem(feed *gtfsparser.Feed, blocks []*Block, max, index int) (blockID, startTime, endTime string) {
	timeType := hhmm
	if index < max {
		blockID = blocks[index].BlockID
		if index > 0 && blocks[index-1].StartAt.Hour == blocks[index].StartAt.Hour {
			timeType = mm
		}
		startTime = Timestamp(blocks[index].StartAt, timeType)
		if index > 0 && blocks[index-1].EndAt.Hour == blocks[index].EndAt.Hour {
			timeType = mm
		}
		endTime = Timestamp(blocks[index].EndAt, timeType)
	} else {
		blockID = ""
		startTime = ""
		endTime = ""
	}
	return blockID, startTime, endTime
}

// BlockCalendarCSV -
func BlockCalendarCSV(feed *gtfsparser.Feed, blocks []*Block, max, index int) (text string) {
	blockID, startTime, endTime := BlockCalendarItem(feed, blocks, max, index)
	text = blockID + "," + startTime + "," + endTime + ","
	return text
}

// func addTripToBlockSchedule(trip *gtfs.Trip, dayOfWeek int) (blockSchedule BlockSchedule) {
// 	blockSchedule = BlockSchedule{}
// 	blockSchedule.Blocktable[dayOfWeek] = blocktable
// 	return blockSchedule
// }

const layoutCSV = "January 2006"

func printBlockMonthCSV(feed *gtfsparser.Feed, filename string, blockCalendar *BlockCalendar, monthStarting gtfs.Date) (err error) {
	for _, feedInfo := range feed.FeedInfos {
		title := feedInfo.Publisher_name + "\n" + "Version:" + feedInfo.Version + "\n"
		for _, agency := range feed.Agencies {
			dayOfWeek := findDaysOfWeekAbbrev(feed, agency)
			weekday := int(toTime(monthStarting).Weekday())
			file, err := os.Create(agency.Name + "-" + filename)
			check(err, "File create error")
			defer file.Close()
			// Get valid date range
			start, end := getFeedDateRange(feed, 0)
			// Print header
			dayHeaderA := ",%s, %d,"
			dayHeaderB := "#,Start,End,"
			lineA := ""
			lineB := ""
			for day := 1; day <= daysThisMonth(); day++ {
				lineA += fmt.Sprintf(dayHeaderA, dayOfWeek[(day+weekday-1)%7], day)
				lineB += dayHeaderB
			}
			lineA += "\n"
			lineB += "\n"
			// year, month, day := time.Now().Date()

			header := fmt.Sprintf("%s\nTransit Block Calendar\nFrom: %s - To: %s\n%s\n", agency.Name, Datestamp(start), Datestamp(end), time.Now().Format(layoutCSV))
			file.WriteString(title + header + lineA + lineB)
			log.Printf(title + header + lineA + lineB)

			tablelength := 0
			for i := 0; i < len(blockCalendar.BlockDays); i++ {
				if blockCalendar.BlockDays[i].Blocks != nil {
					if tablelength < len(blockCalendar.BlockDays[i].Blocks) {
						tablelength = len(blockCalendar.BlockDays[i].Blocks)
					}
				}
			}
			for index := 0; index < tablelength; index++ {
				var output string
				for _, v := range blockCalendar.BlockDays {
					output += BlockCalendarCSV(feed, v.Blocks, len(v.Blocks), index)
				}
				output += "\n"
				_, err = file.WriteString(output)
				log.Printf("%s", output)
				check(err, "File write error")
				file.Sync()
			}
		}
	}
	return err
}

func printBlockWeekCSV(feed *gtfsparser.Feed, filename string, blockSchedule *BlockSchedule, blockID string, weekEnding gtfs.Date) (err error) {
	for _, feedInfo := range feed.FeedInfos {
		title := feedInfo.Publisher_name + "\n" + "Version:" + feedInfo.Version + "\n"
		for _, agency := range feed.Agencies {
			dayOfWeek := findDaysOfWeekAbbrev(feed, agency)
			// Create output CSV file
			file, err := os.Create(agency.Name + "-" + filename)
			check(err, "File create error")
			defer file.Close()
			// Get valid date range
			start, end := getFeedDateRange(feed, 0)
			// Print header
			header := fmt.Sprintf("%s\nTransit Block Schedule\nBlock #%s\nFrom: %s - To: %s\n%s\n%s\n", agency.Name, blockID, Datestamp(start), Datestamp(end), WeekEnding(weekEnding), WeekDatesCSV(weekEnding))
			dayHeader := "#,Trip ID,S,D,%s,"
			output := ""
			for day := time.Monday; day <= time.Saturday; day++ {
				output += fmt.Sprintf(dayHeader, dayOfWeek[day])
			}
			output += fmt.Sprintf(dayHeader, dayOfWeek[time.Sunday]) + "\n"
			file.WriteString(title + header + output)
			log.Printf(title + header + output)

			// sort.Sort(v1.Trips.StopTimes[day])
			tablelength := 0
			for i := 0; i < len(blockSchedule.BlockDays); i++ {
				if blockSchedule.BlockDays[i].Blocks != nil {
					for _, blocks := range blockSchedule.BlockDays[i].Blocks {
						tablelength += len(blocks.Trips)
					}
				}
			}
			for index := 0; index < tablelength; index++ {
				var line [7]string
				var output string
				for weekday := time.Monday; weekday <= time.Saturday; weekday++ {
					if blockSchedule.BlockDays[weekday].Blocks != nil {
						for _, blocks := range blockSchedule.BlockDays[weekday].Blocks {
							line[weekday] = BlockItemCSV(feed, blocks.Trips, len(blocks.Trips), index)
						}
					} else {
						line[weekday] = ",,,,,"
					}
					output += line[weekday]
				}
				if blockSchedule.BlockDays[time.Sunday].Blocks != nil {
					for _, blocks := range blockSchedule.BlockDays[time.Sunday].Blocks {
						trips := blocks.Trips
						line[time.Sunday] = BlockItemCSV(feed, trips, len(trips), index)
						output += line[time.Sunday]
					}
				} else {
					line[time.Sunday] = ",,,,,"
				}
				output += "\n"
				_, err = file.WriteString(output)
				log.Printf("%s", line)
				check(err, "File write error")
				file.Sync()
			}
		}
	}
	return err
}

// Trips -
type Trips []*gtfs.Trip

// Len -
func (s Trips) Len() int { return len(s) }

// Swap -
func (s Trips) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// ByDepartureTime -
type ByDepartureTime struct{ Trips }

// Less -
func (s ByDepartureTime) Less(i, j int) bool {
	return toSeconds(s.Trips[i].StopTimes[0].Departure_time) < toSeconds(s.Trips[j].StopTimes[0].Departure_time)
}

func sortBlockSchedule(blockSchedule BlockSchedule) {
	for weekday := 0; weekday < 7; weekday++ {
		if len(blockSchedule.BlockDays) != 0 {
			if blockSchedule.BlockDays[weekday] != nil {
				for _, blocks := range blockSchedule.BlockDays[weekday].Blocks {
					sort.Sort(ByDepartureTime{blocks.Trips})
				}
			}
		}
	}
}

// Blocks -
type Blocks []*Block

// Len -
func (s Blocks) Len() int { return len(s) }

// Swap -
func (s Blocks) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// ByStartTime -
type ByStartTime struct{ Blocks }

// Less -
func (s ByStartTime) Less(i, j int) bool {
	return toSeconds(s.Blocks[i].StartAt) < toSeconds(s.Blocks[j].StartAt)
}

func sortBlockCalendar(blockCalendar BlockCalendar) {

	for _, day := range blockCalendar.BlockDays {
		if day != nil {
			if day.Blocks != nil {
				for _, block := range day.Blocks {
					if block.Trips != nil {
						sort.Sort(ByDepartureTime{block.Trips})
					}
					block.StartAt = block.Trips[0].StopTimes[0].Departure_time
					lastTrip := block.Trips[len(block.Trips)-1]
					lastArrival := lastTrip.StopTimes[len(lastTrip.StopTimes)-1].Arrival_time
					block.EndAt = lastArrival
				}
				sort.Sort(ByStartTime{day.Blocks})
			}
		}
	}
}

func createWeekSchedule(blocktable *Blocktable, weekEnding gtfs.Date) (blockSchedule BlockSchedule) {
	blockSchedule = BlockSchedule{}
	startOfWeek := toDate(toTime(weekEnding).Add(-time.Hour * 7 * 24).Date())
	for weekday := 0; weekday < 7; weekday++ {
		blockDay := BlockDay{}
		blockDay.Date = DateAdd(startOfWeek, weekday)
		blockSchedule.BlockDays = append(blockSchedule.BlockDays, &blockDay)
	}
	for weekday := 0; weekday < len(ServicesWeek); weekday++ {
		// log.Printf("\nWeekday[%d]:", weekday)
		// output := ""
		for _, servicetable := range blocktable.Servicetables {
			for _, service := range ServicesWeek[weekday] {
				if servicetable.Service.Id == service.Id {
					for _, trip := range servicetable.Trips {
						weekDuration := time.Duration(7-weekday) * 24 * time.Hour
						weekdate := toTime(weekEnding).Add(-weekDuration)
						if weekdate.After(toTime(trip.Service.Start_date)) && weekdate.Before(toTime(trip.Service.End_date)) {
							if blockSchedule.BlockDays[weekday].Blocks == nil {
								block := Block{}
								block.BlockID = blocktable.BlockID
								blockSchedule.BlockDays[weekday].Blocks = append(blockSchedule.BlockDays[weekday].Blocks, &block)
							}
							for _, block := range blockSchedule.BlockDays[weekday].Blocks {
								block.Trips = append(block.Trips, trip)
								// output += fmt.Sprintf("[%s]%s-%s-%s", block.BlockID, trip.Route.Short_name, trip.Service.Id, Timestamp(trip.StopTimes[0].Departure_time, hhmm))
							}
						}
					}
				}
			}
		}
		// log.Println(output)
	}
	return blockSchedule
}

func dayOfWeek(date gtfs.Date) (day int) {
	return int(toTime(date).Weekday())
}

func weeksThisMonth() (weeks int) {
	year, month, _ := time.Now().Date()
	first := dayOfWeek(toDate(year, month, 1))
	last := dayOfWeek(toDate(year, month, daysThisMonth()))
	weeks = (daysThisMonth() + (first + last)) / 7
	return weeks
}

// DateAdd -
func DateAdd(date gtfs.Date, days int) (result gtfs.Date) {
	return toDate(toTime(date).Add(time.Hour * time.Duration(days*24)).Date())
}

func createBlockCalendar(blocktables []*Blocktable) (blockCalendar BlockCalendar) {
	blockCalendar = BlockCalendar{}
	year, month, _ := time.Now().Date()
	for day := 1; day <= daysThisMonth(); day++ {
		blockDay := BlockDay{}
		blockDay.Date = toDate(year, month, day)
		blockCalendar.BlockDays = append(blockCalendar.BlockDays, &blockDay)
	}
	weekStartDate := toDate(year, month, 1)
	firstOfMonth := dayOfWeek(weekStartDate)
	firstOfWeek := firstOfMonth
	weekEndDate := toDate(year, month, 8-firstOfMonth)
	for week := 0; week < weeksThisMonth(); week++ {
		weekdays := 7
		if week == 0 {
			weekdays = 8 - firstOfMonth
			if firstOfMonth == int(time.Sunday) {
				weekdays = 1
			}
		} else if week == weeksThisMonth()-1 {
			weekdays = 7 - int(weekEndDate.Day)
		}
		for _, blocktable := range blocktables {
			weekSchedule := createWeekSchedule(blocktable, weekEndDate)
			for day := 0; day < weekdays; day++ {
				for _, block := range weekSchedule.BlockDays[day+firstOfWeek-1].Blocks {
					weekStartDay := int(weekStartDate.Day) - 1
					if week == 3 {
						log.Println(day, weekStartDay, block.BlockID)
					}
					blockCalendar.BlockDays[day+weekStartDay].Blocks = append(blockCalendar.BlockDays[day+weekStartDay].Blocks, block)
				}
			}
		}
		weekEndDate = DateAdd(weekEndDate, 7)
		weekStartDate = DateAdd(weekStartDate, weekdays)
		firstOfWeek = 1
	}
	return blockCalendar
}

func createBlockSchedule(blocktables []*Blocktable, blockID string, weekEnding gtfs.Date) (blockSchedule BlockSchedule) {
	blockSchedule = BlockSchedule{}
	for _, blocktable := range blocktables {
		if blocktable.BlockID == blockID {
			blockSchedule = createWeekSchedule(blocktable, weekEnding)
			break
		}
	}
	return blockSchedule
}

// DeadheadDay -
type DeadheadDay struct {
	Date  gtfs.Date
	Trips []*gtfs.Trip
}

// DeadheadSchedule -
type DeadheadSchedule struct {
	DeadheadDays []*DeadheadDay //	Blocks [7]*Block
}

// StopType -
type StopType int

const (
	// Regular - 0 or empty - Regularly scheduled pickup or dropoff
	Regular StopType = iota
	// NoService - 1 - No pickup or dropoff available.
	NoService
	// AgencyConfirm - 2 - Must phone agency to arrange pickup or dropoff.
	AgencyConfirm
	// DriverCoordinate - 3 - Must coordinate with driver to arrange pickup or dropoff.
	DriverCoordinate
)

// DeadheadItem -
func DeadheadItem(feed *gtfsparser.Feed, trips []*gtfs.Trip, max, index int) (blockID, routeName, tripID, serviceID, scheduleTime, directionID string) {
	var timeType TimeType
	if index > 0 && index < max && trips[index-1].StopTimes[0].Departure_time.Hour == trips[index].StopTimes[0].Departure_time.Hour {
		timeType = mm
	} else {
		timeType = hhmm
	}

	if index < max {
		routeName = trips[index].Route.Short_name
		if len(routeName) == 0 {
			routeName = trips[index].Short_name
		}
		if len(routeName) == 0 {
			routeName = firstWords(trips[index].Headsign, 1)
		}
		scheduleTime = Timestamp(trips[index].StopTimes[0].Departure_time, timeType)
		directionID = strconv.Itoa(int(trips[index].Direction_id))
		serviceID = trips[index].Service.Id
		blockID = trips[index].Block_id
		tripID = trips[index].Id
	} else {
		tripID = ""
		blockID = ""
		routeName = ""
		serviceID = ""
		directionID = ""
		scheduleTime = ""
	}
	return blockID, routeName, tripID, serviceID, scheduleTime, directionID
}

// DeadheadItemCSV -
func DeadheadItemCSV(feed *gtfsparser.Feed, trips []*gtfs.Trip, max, index int) (text string) {
	blockID, routeName, tripID, serviceID, startTime, directionID := DeadheadItem(feed, trips, max, index)
	text = blockID + "," + routeName + "," + tripID + "," + serviceID + "," + directionID + "," + startTime + ","
	return text
}

func createDeadheadSchedule(blocktables []*Blocktable, weekEnding gtfs.Date) (deadheadSchedule DeadheadSchedule) {
	deadheadSchedule = DeadheadSchedule{}
	startOfWeek := toDate(toTime(weekEnding).Add(-time.Hour * 7 * 24).Date())
	for weekday := 0; weekday < 7; weekday++ {
		deadheadDay := DeadheadDay{}
		deadheadDay.Date = DateAdd(startOfWeek, weekday)
		deadheadSchedule.DeadheadDays = append(deadheadSchedule.DeadheadDays, &deadheadDay)
	}
	for weekday := 0; weekday < len(ServicesWeek); weekday++ {
		// log.Printf("\nWeekday[%d]:", weekday)
		// output := ""
		for _, blocktable := range blocktables {
			for _, servicetable := range blocktable.Servicetables {
				for _, service := range ServicesWeek[weekday] {
					if servicetable.Service.Id == service.Id {
						for _, trip := range servicetable.Trips {
							weekDuration := time.Duration(7-weekday) * 24 * time.Hour
							weekdate := toTime(weekEnding).Add(-weekDuration)
							if weekdate.After(toTime(trip.Service.Start_date)) && weekdate.Before(toTime(trip.Service.End_date)) {
								trips := deadheadSchedule.DeadheadDays[weekday].Trips
								var servicedStops, unservicedStops int
								for _, stopTime := range trip.StopTimes {
									if StopType(stopTime.Pickup_type) == NoService && StopType(stopTime.Drop_off_type) == NoService {
										unservicedStops++
										log.Println(stopTime.Pickup_type, ": No Service [", unservicedStops, "]", stopTime.Stop.Code, stopTime.Sequence, Timestamp(stopTime.Arrival_time, hhmm), trip.Route.Id, trip.Id)
									} else if StopType(stopTime.Pickup_type) == NoService {
										log.Println(stopTime.Pickup_type, ": No Pickup: [", unservicedStops, "]", stopTime.Stop.Code, stopTime.Sequence, Timestamp(stopTime.Arrival_time, hhmm), trip.Route.Id, trip.Id)
									} else if StopType(stopTime.Drop_off_type) == NoService {
										log.Println(stopTime.Pickup_type, ": No Dropoff: [", unservicedStops, "]", stopTime.Stop.Code, stopTime.Sequence, Timestamp(stopTime.Arrival_time, hhmm), trip.Route.Id, trip.Id)
									} else {
										servicedStops++
									}
								}
								if unservicedStops > 0 { // && servicedStops == 0
									log.Println(": No Service [", weekday, "][", trip.Route.Id, trip.Id, unservicedStops, servicedStops, "]")
									trips = append(trips, trip)
								}
							}
						}
					}
				}
			}
		}
	}
	return deadheadSchedule
}

func printDeadheadWeekCSV(feed *gtfsparser.Feed, filename string, deadheadSchedule *DeadheadSchedule, weekEnding gtfs.Date) (err error) {
	for _, feedInfo := range feed.FeedInfos {
		title := feedInfo.Publisher_name + "\n" + "Version:" + feedInfo.Version + "\n"
		for _, agency := range feed.Agencies {
			dayOfWeek := findDaysOfWeekAbbrev(feed, agency)
			// Create output CSV file
			file, err := os.Create(agency.Name + "-" + filename)
			check(err, "File create error")
			defer file.Close()
			// Get valid date range
			start, end := getFeedDateRange(feed, 0)
			// Print header
			header := fmt.Sprintf("%s\nTransit Deadhead Schedule\nFrom: %s - To: %s\n%s\n%s\n", agency.Name, Datestamp(start), Datestamp(end), WeekEnding(weekEnding), WeekDatesCSV(weekEnding))
			dayHeader := "#,Trip ID,S,D,%s,"
			output := ""
			for day := time.Monday; day <= time.Saturday; day++ {
				output += fmt.Sprintf(dayHeader, dayOfWeek[day])
			}
			output += fmt.Sprintf(dayHeader, dayOfWeek[time.Sunday]) + "\n"
			file.WriteString(title + header + output)
			log.Printf(title + header + output)

			tablelength := 0
			for i := 0; i < len(deadheadSchedule.DeadheadDays); i++ {
				trips := deadheadSchedule.DeadheadDays[i].Trips
				if trips != nil {
					if tablelength < len(trips) {
						tablelength = len(trips)
					}
				}
			}

			for index := 0; index < tablelength; index++ {
				var line [7]string
				var output string
				for weekday := time.Monday; weekday <= time.Saturday; weekday++ {
					if deadheadSchedule.DeadheadDays[weekday].Trips != nil {
						trips := deadheadSchedule.DeadheadDays[weekday].Trips
						line[weekday] = BlockItemCSV(feed, trips, len(trips), index)
					} else {
						line[weekday] = ",,,,,"
					}
					output += line[weekday]
				}
				if deadheadSchedule.DeadheadDays[time.Sunday].Trips != nil {
					trips := deadheadSchedule.DeadheadDays[time.Sunday].Trips
					line[time.Sunday] = DeadheadItemCSV(feed, trips, len(trips), index)
					output += line[time.Sunday]
				} else {
					line[time.Sunday] = ",,,,,"
				}
				output += "\n"
				_, err = file.WriteString(output)
				log.Printf("%s", line)
				check(err, "File write error")
				file.Sync()
			}
		}
	}
	return err
}

// StopPoint -
type StopPoint struct {
	StopID           string
	StopCode         string
	StopName         string
	StopDescription  string
	StopSequence     int
	IsTimingPoint    bool
	Lat              float32
	Lng              float32
	DistanceTraveled float32
}

func findStopPointsForTrip(feed *gtfsparser.Feed, tripID string, timingPoint bool) (stopPoints []*StopPoint) {
	for _, trip := range feed.Trips {
		if trip.Id == tripID {
			for _, stopTime := range trip.StopTimes {
				stopPoint := StopPoint{}
				stopPoint.StopID = stopTime.Stop.Id
				stopPoint.StopCode = stopTime.Stop.Code
				stopPoint.StopName = stopTime.Stop.Name
				stopPoint.StopDescription = stopTime.Stop.Desc
				stopPoint.StopSequence = stopTime.Sequence
				stopPoint.IsTimingPoint = stopTime.Timepoint
				stopPoint.Lat = stopTime.Stop.Lat
				stopPoint.Lat = stopTime.Stop.Lat
				stopPoint.DistanceTraveled = stopTime.Shape_dist_traveled
				if timingPoint == true {
					if stopPoint.IsTimingPoint == true {
						stopPoints = append(stopPoints, &stopPoint)
					}
				} else {
					stopPoints = append(stopPoints, &stopPoint)
				}
			}
		}
	}
	return stopPoints
}

func test(data []byte, err error) {

}

func GetStopPointsForTripJSON(feed *gtfsparser.Feed, tripID string, timingPoint bool) (stops map[string]StopPoint) {
	stopPoints := findStopPointsForTrip(feed, tripID, timingPoint)
	data, err := json.Marshal(stopPoints)
	test(data, err)
	return stops
}

func SetAuthority(authority string) (authorities map[string]authority) {

}

func SetAgency(agency string) {

}

func GetTrip(tripID string) {

}

func GetStop(stopCode string) {

}
