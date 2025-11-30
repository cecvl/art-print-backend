package processing

import (
	"context"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	vision "cloud.google.com/go/vision/apiv1"
	"github.com/cecvl/art-print-backend/internal/firebase"
	visionpb "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

// StartWorker polls the Firestore processing_queue for pending jobs and processes them.
// This is intentionally minimal: it runs SafeSearch on the image URL and writes results
// to the artwork document under `analysis.safeSearch`.
func StartWorker(ctx context.Context) error {
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	fs := firebase.FirestoreClient

	log.Println("▶️ Processing worker started")
	for {
		// find pending jobs
		q := fs.Collection("processing_queue").Where("status", "==", "pending").Limit(5)
		snaps, err := q.Documents(ctx).GetAll()
		if err != nil {
			log.Printf("⚠️ Failed to query processing_queue: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if len(snaps) == 0 {
			time.Sleep(2 * time.Second)
			continue
		}

		for _, s := range snaps {
			go func(s *firestore.DocumentSnapshot) {
				jobRef := s.Ref
				jobData := s.Data()
				artworkId, _ := jobData["artworkId"].(string)
				frameId, _ := jobData["frameId"].(string)
				cloudInfo, _ := jobData["cloudinary"].(map[string]interface{})
				imgUrl, _ := cloudInfo["secureUrl"].(string)

				// mark job processing
				jobRef.Update(ctx, []firestore.Update{{Path: "status", Value: "processing"}, {Path: "startedAt", Value: time.Now()}})

				// run SafeSearch
				visImg := vision.NewImageFromURI(imgUrl)
				res, err := client.DetectSafeSearch(ctx, visImg, nil)
				if err != nil {
					log.Printf("❌ SafeSearch failed for artwork %s: %v", artworkId, err)
					jobRef.Update(ctx, []firestore.Update{{Path: "status", Value: "failed"}, {Path: "error", Value: err.Error()}})
					return
				}

				// run WebDetection (copyright / similar images)
				webRes, wErr := client.DetectWeb(ctx, visImg, nil)
				if wErr != nil {
					log.Printf("⚠️ WebDetection failed for job (artwork:%s frame:%s) %v", artworkId, frameId, wErr)
				}
				// prepare web entities summary
				var webEntities []map[string]interface{}
				if webRes != nil && webRes.WebEntities != nil {
					for _, we := range webRes.WebEntities {
						if we == nil {
							continue
						}
						webEntities = append(webEntities, map[string]interface{}{
							"entityId":    we.EntityId,
							"score":       we.Score,
							"description": we.Description,
						})
					}
				}

				// fetch image and run local analysis (blur, color depth, dimensions)
				img, format, err := fetchImage(imgUrl)
				if err != nil {
					log.Printf("❌ Failed to fetch image for artwork %s: %v", artworkId, err)
					jobRef.Update(ctx, []firestore.Update{{Path: "status", Value: "failed"}, {Path: "error", Value: err.Error()}})
					return
				}

				w := img.Bounds().Dx()
				h := img.Bounds().Dy()

				// compute blur score
				blurScore := ComputeBlurFromImage(img)
				colorDepth := DetectColorDepth(img)

				// Persist analysis (shared)
				analysis := map[string]interface{}{
					"safeSearch": map[string]interface{}{
						"adult":    res.Adult.String(),
						"violence": res.Violence.String(),
						"racy":     res.Racy.String(),
						"medical":  res.Medical.String(),
						"spoof":    res.Spoof.String(),
					},
					"format":     format,
					"width":      w,
					"height":     h,
					"blurScore":  blurScore,
					"colorDepth": colorDepth,
					"checkedAt":  time.Now(),
				}

				// Determine overall processing status: fail if NSFW detected
				procStatus := "ready"
				procErrors := []string{}
				if res.Adult == visionpb.Likelihood_LIKELY || res.Adult == visionpb.Likelihood_VERY_LIKELY {
					procStatus = "failed"
					procErrors = append(procErrors, "nsfw_adult")
				}
				if res.Violence == visionpb.Likelihood_LIKELY || res.Violence == visionpb.Likelihood_VERY_LIKELY {
					procStatus = "failed"
					procErrors = append(procErrors, "nsfw_violence")
				}

				// If this is a frame job, run label detection to verify it's a frame
				if frameId != "" {
					labels, lErr := client.DetectLabels(ctx, visImg, nil, 10)
					var labelSumm []map[string]interface{}
					foundFrame := false
					if lErr == nil {
						for _, lb := range labels {
							if lb == nil {
								continue
							}
							desc := lb.Description
							labelSumm = append(labelSumm, map[string]interface{}{"description": desc, "score": lb.Score})
							if strings.Contains(strings.ToLower(desc), "frame") || strings.Contains(strings.ToLower(desc), "picture frame") {
								foundFrame = true
							}
						}
					} else {
						log.Printf("⚠️ Label detection failed for frame %s: %v", frameId, lErr)
					}
					analysis["labels"] = labelSumm
					if !foundFrame {
						procStatus = "failed"
						procErrors = append(procErrors, "not_a_frame")
					}

					// Persist to frames doc
					frameRef := firebase.FirestoreClient.Collection("frames").Doc(frameId)
					_, err = frameRef.Set(ctx, map[string]interface{}{"analysis": analysis, "processingStatus": procStatus, "processingErrors": procErrors}, firestore.MergeAll)
					if err != nil {
						log.Printf("❌ Failed to update frame %s after analysis: %v", frameId, err)
						jobRef.Update(ctx, []firestore.Update{{Path: "status", Value: "failed"}, {Path: "error", Value: err.Error()}})
						return
					}

					jobRef.Update(ctx, []firestore.Update{{Path: "status", Value: "done"}, {Path: "finishedAt", Value: time.Now()}})
					log.Printf("✅ Processed frame %s (job %s) - status=%s", frameId, jobRef.ID, procStatus)
					return
				}

				// Persist results to artwork doc
				artRef := firebase.FirestoreClient.Collection("artworks").Doc(artworkId)
				_, err = artRef.Set(ctx, map[string]interface{}{"analysis": analysis, "processingStatus": procStatus, "processingErrors": procErrors}, firestore.MergeAll)
				if err != nil {
					log.Printf("❌ Failed to update artwork %s after analysis: %v", artworkId, err)
					jobRef.Update(ctx, []firestore.Update{{Path: "status", Value: "failed"}, {Path: "error", Value: err.Error()}})
					return
				}

				jobRef.Update(ctx, []firestore.Update{{Path: "status", Value: "done"}, {Path: "finishedAt", Value: time.Now()}})
				log.Printf("✅ Processed artwork %s (job %s) - status=%s", artworkId, jobRef.ID, procStatus)
			}(s)
		}

		// small sleep to permit other loops
		time.Sleep(500 * time.Millisecond)
	}
}
