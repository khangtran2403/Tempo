package connector

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brokeboycoding/tempo/internal/storage"
	"github.com/brokeboycoding/tempo/pkg/crypto"
	"github.com/jomei/notionapi"
)

type NotionConnector struct {
	integrationRepo storage.IntegrationRepository
	encryptionKey   string
}

func NewNotionConnector(
	integrationRepo storage.IntegrationRepository,
	encryptionKey string,
) *NotionConnector {
	return &NotionConnector{
		integrationRepo: integrationRepo,
		encryptionKey:   encryptionKey,
	}
}

func (n *NotionConnector) Name() string {
	return "notion"
}

func (n *NotionConnector) ValidateConfig(cfg map[string]interface{}) error {
	if _, ok := cfg["database_id"].(string); !ok {
		return fmt.Errorf("notion: 'database_id' is required")
	}
	if _, ok := cfg["properties"].(string); !ok {
		return fmt.Errorf("notion: 'properties' is required and must be a JSON string")
	}
	return nil
}

func (n *NotionConnector) Execute(ctx context.Context, config map[string]interface{}, prevResults map[string]interface{}) (interface{}, error) {
	action, _ := config["action"].(string)
	if action == "" {
		action = "create_page"
	}

	switch action {
	case "create_page":
		return n.createPage(ctx, config, prevResults)
	default:
		return nil, fmt.Errorf("notion: unknown action '%s'", action)
	}
}

func (n *NotionConnector) createPage(ctx context.Context, config map[string]interface{}, prevResults map[string]interface{}) (interface{}, error) {
	integrationID, ok := config["integration_id"].(string)
	if !ok || integrationID == "" {
		return nil, fmt.Errorf("notion: 'integration_id' is required")
	}
	client, err := n.getNotionClient(ctx, integrationID)
	if err != nil {
		return nil, err
	}

	databaseIDStr, _ := config["database_id"].(string)
	
	// The properties and content strings have already been rendered by the activity
	renderedPropsStr, _ := config["properties"].(string)
	renderedContent, _ := config["content"].(string)

	var properties notionapi.Properties
	if err := json.Unmarshal([]byte(renderedPropsStr), &properties); err != nil {
		return nil, fmt.Errorf("notion: failed to unmarshal properties JSON: %w. Make sure it's valid JSON matching Notion's API structure", err)
	}

	request := notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: notionapi.DatabaseID(databaseIDStr),
		},
		Properties: properties,
	}

	if renderedContent != "" {
		blocks, err := markdownToBlocks(renderedContent)
		if err != nil {
			return nil, fmt.Errorf("notion: failed to parse content: %w", err)
		}
		request.Children = blocks
	}
	
	page, err := client.Page.Create(ctx, &request)
	if err != nil {
		return nil, fmt.Errorf("notion: failed to create page: %w", err)
	}

	return page, nil
}
func (n *NotionConnector) getNotionClient(ctx context.Context, integrationID string) (*notionapi.Client, error) {
	integration, err := n.integrationRepo.GetByID(ctx, integrationID)
	if err != nil || integration == nil {
		return nil, fmt.Errorf("notion: integration with ID '%s' not found", integrationID)
	}
	if integration.Provider != "notion" {
		return nil, fmt.Errorf("notion: integration '%s' is not a Notion integration", integrationID)
	}

	decryptedToken, err := crypto.Decrypt(integration.AccessToken, n.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("notion: failed to decrypt token: %w", err)
	}

	client := notionapi.NewClient(notionapi.Token(decryptedToken))
	return client, nil
}

func markdownToBlocks(md string) ([]notionapi.Block, error) {
	var blocks []notionapi.Block
	if md != "" {
		blocks = append(blocks, &notionapi.ParagraphBlock{
			BasicBlock: notionapi.BasicBlock{
				Object: notionapi.ObjectTypeBlock,
				Type:   notionapi.BlockTypeParagraph,
			},
			Paragraph: notionapi.Paragraph{
				RichText: []notionapi.RichText{
					{
						Type: "text",
						Text: &notionapi.Text{
							Content: md,
						},
					},
				},
			},
		})
	}
	return blocks, nil
}
