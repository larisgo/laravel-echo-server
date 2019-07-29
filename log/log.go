package log

import (
	"fmt"
	"github.com/gookit/color"
	"log"
)

/**
 * Console log line.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func Comment(message ...interface{}) {
	fmt.Print(color.Tag("default").Sprint(fmt.Sprint(message...)))
}

/**
 * Console log line.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func Line(message ...interface{}) {
	fmt.Println(color.Tag("default").Sprint(fmt.Sprint(message...)))
}

/**
 * Console log default.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func Default(message ...interface{}) {
	log.Println(color.Tag("default").Sprint(fmt.Sprint(message...)))
}

/**
 * Console log info.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func Info(message ...interface{}) {
	log.Println(color.Info.Sprint(fmt.Sprint(message...)))
}

/**
 * Console log success.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func Success(message ...interface{}) {
	log.Println(color.Success.Sprint(fmt.Sprint(message...)))
}

/**
 * Console log info.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func Error(message ...interface{}) {
	log.Println(color.Danger.Sprint(fmt.Sprint(message...)))
}

/**
 * Console log warning.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func Warning(message ...interface{}) {
	log.Println(color.Warn.Sprint(fmt.Sprint(message...)))
}

/**
 * Console log fatal.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func Fatal(message ...interface{}) {
	log.Fatal(color.Error.Sprint(fmt.Sprint(message...)))
}
