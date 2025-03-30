package bot

// func createTestBot(t *testing.T) (*bot, *tg_api_mock.MockTgAPI) {
// 	ctrl := gomock.NewController(t)
// 	tgAPI := tg_api_mock.NewMockTgAPI(ctrl)

// 	b := &bot{
// 		tgAPI: tgAPI,
// 		chats: make(map[int64]*chat),
// 		cfg: &config.Config{
// 			Greeting:  "Привет!",
// 			Questions: []string{"Ты любишь лето?", "Ты любишь весну?"},
// 			Calculation: map[string]*config.PsyhoType{
// 				"Любит лето": &config.PsyhoType{
// 					Description:     "радостный",
// 					QuestionIndexes: []int{0},
// 				},
// 				"Любит весну": &config.PsyhoType{
// 					Description:     "мечтательный",
// 					QuestionIndexes: []int{1},
// 				},
// 			},
// 			TelegramInfo: &config.TelegramInfo{
// 				Token:      "abc",
// 				WebhookUrl: "localhost:800",
// 			},
// 		},
// 	}

// 	return b, tgAPI
// }

// func Test_startQuiz(t *testing.T) {
// 	b, api := createTestBot(t)
// 	api.EXPECT().SendMessage(int64(123), tgbotapi.MessageConfig{
// 		Text: "Привет!",
// 		BaseChat: tgbotapi.BaseChat{
// 			ChatID: 123,
// 			ReplyMarkup: tgbotapi.InlineKeyboardMarkup{
// 				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
// 					{
// 						tgbotapi.InlineKeyboardButton{
// 							Text:         "Начать тест",
// 							CallbackData: &startCmd,
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}).Return(nil)
// 	err := b.startQuiz(&tgbotapi.Message{
// 		Chat: &tgbotapi.Chat{
// 			ID: 123,
// 		},
// 	})
// 	assert.NoError(t, err)
// }

// func Test_sendQuestion(t *testing.T) {
// 	b, api := createTestBot(t)
// 	b.chats[123] = &chat{
// 		id:            123,
// 		questionIndex: 0,
// 		answers:       []int{},
// 	}
// 	s1 := "0:Нет"
// 	s2 := "0:Частично"
// 	s3 := "0:Да"
// 	api.EXPECT().SendMessage(int64(123), tgbotapi.MessageConfig{
// 		Text: formatQuestion(0, "Ты любишь лето?"),
// 		BaseChat: tgbotapi.BaseChat{
// 			ChatID: 123,
// 			ReplyMarkup: tgbotapi.InlineKeyboardMarkup{
// 				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
// 					{
// 						tgbotapi.InlineKeyboardButton{
// 							Text:         "Нет",
// 							CallbackData: &s1,
// 						},
// 					},
// 					{
// 						tgbotapi.InlineKeyboardButton{
// 							Text:         "Частично",
// 							CallbackData: &s2,
// 						},
// 					},
// 					{
// 						tgbotapi.InlineKeyboardButton{
// 							Text:         "Да",
// 							CallbackData: &s3,
// 						},
// 					},
// 					{
// 						tgbotapi.InlineKeyboardButton{
// 							Text:         "Сбросить результаты и начать заново",
// 							CallbackData: &startCmd,
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}).Return(nil)
// 	err := b.sendQuestion(123)
// 	assert.NoError(t, err)
// }

// func Test_sendResults(t *testing.T) {
// 	b, api := createTestBot(t)
// 	b.chats[123] = &chat{
// 		id:            123,
// 		questionIndex: 1,
// 		answers:       []int{1, 3},
// 	}
// 	api.EXPECT().SendMessage(int64(123), tgbotapi.MessageConfig{
// 		BaseChat: tgbotapi.BaseChat{
// 			ChatID: 123,
// 		},
// 		Text: "Результат:\nВаш тип личности - Любит весну\nОписание типа: мечтательный",
// 	})
// 	err := b.sendResults(123)
// 	assert.NoError(t, err)
// }

// func Test_handleCallback(t *testing.T) {
// 	b, api := createTestBot(t)
// 	b.chats[123] = &chat{
// 		id:            123,
// 		questionIndex: 1,
// 		answers:       []int{1, 3},
// 	}
// 	s1 := "0:Нет"
// 	s2 := "0:Частично"
// 	s3 := "0:Да"
// 	api.EXPECT().SendMessage(int64(123), tgbotapi.MessageConfig{
// 		Text: formatQuestion(0, "Ты любишь лето?"),
// 		BaseChat: tgbotapi.BaseChat{
// 			ChatID: 123,
// 			ReplyMarkup: tgbotapi.InlineKeyboardMarkup{
// 				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
// 					{
// 						tgbotapi.InlineKeyboardButton{
// 							Text:         "Нет",
// 							CallbackData: &s1,
// 						},
// 					},
// 					{
// 						tgbotapi.InlineKeyboardButton{
// 							Text:         "Частично",
// 							CallbackData: &s2,
// 						},
// 					},
// 					{
// 						tgbotapi.InlineKeyboardButton{
// 							Text:         "Да",
// 							CallbackData: &s3,
// 						},
// 					},
// 					{
// 						tgbotapi.InlineKeyboardButton{
// 							Text:         "Сбросить результаты и начать заново",
// 							CallbackData: &startCmd,
// 						},
// 					},
// 				},
// 			},
// 		},
// 	},
// 	)
// 	err := b.handleCallback(&tgbotapi.CallbackQuery{
// 		Message: &tgbotapi.Message{
// 			Chat: &tgbotapi.Chat{
// 				ID: 123,
// 			},
// 		},
// 		Data: "0:Да",
// 	})
// 	assert.NoError(t, err)
// }
