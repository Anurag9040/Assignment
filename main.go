package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()

	app.Use(logger.New())
	var wg sync.WaitGroup
	ch := make(chan bool)

	app.Post("/", func(c *fiber.Ctx) error {
		wg.Add(1)
		go SendRequest(c,&wg,ch)
	     return nil
	})


	go func() {
		wg.Wait()    
		close(ch) 
	}()

	app.Listen(":3000")
}

func SendRequest(c *fiber.Ctx, wg *sync.WaitGroup, ch chan<- bool){

	defer wg.Done()

	var inputJson map[string]interface{}

		if err := c.BodyParser(&inputJson); err != nil {
			 c.Status(fiber.StatusBadRequest).SendString("Error decoding JSON")
			 return 
		}

		outputMap := map[string]interface{}{
			"event":            inputJson["ev"],
			"event_type":       inputJson["et"],
			"app_id":           inputJson["id"],
			"user_id":          inputJson["uid"],
			"message_id":       inputJson["mid"],
			"page_title":       inputJson["t"],
			"page_url":         inputJson["p"],
			"browser_language": inputJson["l"],
			"screen_size":      inputJson["sc"],
		}

		attributes := make(map[string]interface{})
		for key, value := range inputJson {
			if len(key) > 4 && key[:4] == "atrk" {
				attrKey := key[4:]
				attrvalue := fmt.Sprint(value)
				attrValue, attrType := inputJson["atrv"+attrKey], inputJson["atrt"+attrKey]
				attributes[attrvalue] = map[string]interface{}{
					"type":  attrType,
					"value": attrValue,
				}
			}
		}
		outputMap["attributes"] = attributes

		traits := make(map[string]interface{})
		for key, value := range inputJson {
			if len(key) > 5 && key[:5] == "uatrk" {
				traitKey := key[5:]
				traitValue, traitType := inputJson["uatrv"+traitKey], inputJson["uatrt"+traitKey]
				traits[fmt.Sprint(value)] = map[string]interface{}{
					"type":  traitType,
					"value": traitValue,
				}
			}
		}
		outputMap["traits"] = traits

		outputJSON, err := json.Marshal(outputMap)
		if err != nil {
			 c.Status(fiber.StatusInternalServerError).SendString("Error Converting OutputMap To Json")
			 return
		}

		body := bytes.NewBuffer(outputJSON)

		webhookURL := "https://webhook.site/62848008-375c-4ee0-af67-b5ee7ebdfe42"

		resp, err := http.Post(webhookURL, "application/json", body)
		if err != nil {
			 c.Status(fiber.StatusInternalServerError).SendString("Error sending data to webhook.site")
			 return
		}
		defer resp.Body.Close()

}