package rds

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/tailwarden/komiser/models"
	"github.com/tailwarden/komiser/providers"
	"github.com/tailwarden/komiser/utils"
)

func Instances(ctx context.Context, client providers.ProviderClient) ([]models.Resource, error) {
	var config rds.DescribeDBInstancesInput
	resources := make([]models.Resource, 0)
	rdsClient := rds.NewFromConfig(*client.AWSClient)

	oldRegion := client.AWSClient.Region
	client.AWSClient.Region = "us-east-1"
	pricingClient := pricing.NewFromConfig(*client.AWSClient)
	client.AWSClient.Region = oldRegion

	for {
		output, err := rdsClient.DescribeDBInstances(ctx, &config)
		if err != nil {
			return resources, err
		}

		for _, instance := range output.DBInstances {
			tags := make([]models.Tag, 0)
			for _, tag := range instance.TagList {
				tags = append(tags, models.Tag{
					Key:   *tag.Key,
					Value: *tag.Value,
				})
			}

			var _instanceName string
			if instance.DBName == nil {
				_instanceName = *instance.DBInstanceIdentifier
			} else {
				_instanceName = *instance.DBName
			}

			startOfMonth := utils.BeginningOfMonth(time.Now())
			hourlyUsage := 0
			if (*instance.InstanceCreateTime).Before(startOfMonth) {
				hourlyUsage = int(time.Since(startOfMonth).Hours())
			} else {
				hourlyUsage = int(time.Since(*instance.InstanceCreateTime).Hours())
			}

			pricingOutput, err := pricingClient.GetProducts(ctx, &pricing.GetProductsInput{
				ServiceCode: aws.String("AmazonRDS"),
				Filters: []types.Filter{
					{
						Field: aws.String("instanceType"),
						Value: aws.String(*instance.DBInstanceClass),
						Type:  types.FilterTypeTermMatch,
					},
					{
						Field: aws.String("regionCode"),
						Value: aws.String(client.AWSClient.Region),
						Type:  types.FilterTypeTermMatch,
					},
					{
						Field: aws.String("databaseEngine"),
						Value: aws.String(*instance.Engine),
						Type:  types.FilterTypeTermMatch,
					},
				},
				MaxResults: aws.Int32(1),
			})
			if err != nil {
				log.Warnf("Couldn't fetch invocations metric for %s", _instanceName)
			}

			hourlyCost := 0.0
			if pricingOutput != nil && len(pricingOutput.PriceList) > 0 {
				b, _ := json.Marshal(pricingOutput.PriceList[0])
				s, _ := strconv.Unquote(string(b))

				pricingResult := models.PricingResult{}
				err = json.Unmarshal([]byte(s), &pricingResult)
				if err != nil {
					log.WithError(err).Error("could not unmarshal")
				}

				hourlyCostRaw := pricingResult.Terms["OnDemand"][fmt.Sprintf("%s.JRTCKXETXF", pricingResult.Product.Sku)]["priceDimensions"][fmt.Sprintf("%s.JRTCKXETXF.6YS6EN2CT7", pricingResult.Product.Sku)].PricePerUnit.USD
				hourlyCost, _ = strconv.ParseFloat(hourlyCostRaw, 64)
			}

			monthlyCost := float64(hourlyUsage) * hourlyCost

			resources = append(resources, models.Resource{
				Provider:   "AWS",
				Account:    client.Name,
				Service:    "RDS Instance",
				Region:     client.AWSClient.Region,
				ResourceId: *instance.DBInstanceArn,
				Cost:       monthlyCost,
				Name:       _instanceName,
				FetchedAt:  time.Now(),
				Tags:       tags,
				Link:       fmt.Sprintf("https:/%s.console.aws.amazon.com/rds/home?region=%s#database:id=%s", client.AWSClient.Region, client.AWSClient.Region, *instance.DBInstanceIdentifier),
			})
		}

		if aws.ToString(output.Marker) == "" {
			break
		}

		config.Marker = output.Marker
	}
	log.WithFields(log.Fields{
		"provider":  "AWS",
		"account":   client.Name,
		"region":    client.AWSClient.Region,
		"service":   "RDS Instance",
		"resources": len(resources),
	}).Info("Fetched resources")
	return resources, nil
}
