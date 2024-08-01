package ai_app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tiktoken-go/tokenizer"
	"github.com/tmc/langchaingo/textsplitter"

	"tg_ai_service/internal/ai_model"
	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
)

const (
	MAX_TOKENS_PER_CHUNK = 1500
)

type TranslateApp struct {
	ai              common.AI
	defaultModel    string
	chunksInContext int
	maxToken        int
}

func NewTranslateApp(cfg *common.TranslateAppConfig, aiS *ai_model.AiService) (*TranslateApp, error) {
	ai := aiS.GetAiByType(cfg.AiT)
	chunksInContext := cfg.ChunksInContext

	return &TranslateApp{
		ai:              ai,
		defaultModel:    cfg.Model,
		chunksInContext: chunksInContext,
		maxToken:        cfg.MaxToken,
	}, nil
}

func (t *TranslateApp) oneChunkInitialTranslation(sourceLang, targetLang, sourceText string, stepRecord *common.TranslateStepRecord, ctx context.Context) (translation string, err error) {
	systemMessage := fmt.Sprintf("You are an expert linguist, specializing in translation from %s to %s.", sourceLang, targetLang)

	translationPrompt := fmt.Sprintf("This is an %s to %s translation, please provide the %s translation for this text. Do not provide any explanations or text apart from the translation.\n%s: %s\n\n%s:", sourceLang, targetLang, targetLang, sourceLang, sourceText, targetLang)

	start := time.Now()
	defer func() {
		if err == nil {
			inputToken, _ := numTokensInString(translationPrompt, "cl100k_base")
			oneStep := &common.OneAiRecord{
				Cost:       time.Since(start).Milliseconds(),
				InputToken: inputToken,
			}
			oneStep.OutputToken, _ = numTokensInString(translation, "cl100k_base")
			stepRecord.Records = append(stepRecord.Records, oneStep)
			return
		}
		log.Errorf("oneChunkInitialTranslation error: %v", err)
	}()

	translation, err = t.ai.GetCompletion(translationPrompt, systemMessage, t.defaultModel, 0.3, false, ctx)
	if err != nil {
		return "", err
	}
	return translation, nil
}

func (t *TranslateApp) oneChunkReflectOnTranslation(sourceLang, targetLang, sourceText, translation1, country string, stepRecord *common.TranslateStepRecord, ctx context.Context) (reflection string, err error) {
	systemMessage := fmt.Sprintf("You are an expert linguist specializing in translation from %s to %s. You will be provided with a source text and its translation and your goal is to improve the translation.", sourceLang, targetLang)

	var reflectionPrompt string
	if country != "" {
		reflectionPrompt = fmt.Sprintf(`Your task is to carefully read a source text and a translation from %s to %s, and then give constructive criticism a
		helpful suggestions to improve the translation. The final style and tone of the translation should match the style of %s colloquially spoken in %s.\n\nThe
		source text and initial translation, delimited by XML tags <SOURCE_TEXT></SOURCE_TEXT> and <TRANSLATION></TRANSLATION>, are as
	follows:\n\n<SOURCE_TEXT>\n%s\n</SOURCE_TEXT>\n\n<TRANSLATION>\n%s\n</TRANSLATION>\n\nWhen writing suggestions, pay attention to whether there are ways to
		improve the translation's (i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text), (ii) fluency (by applying %s
		grammar, spelling and punctuation rules, and ensuring there are no unnecessary repetitions), (iii) style (by ensuring the translations reflect the style of
		the source text and take into account any cultural context), (iv) terminology (by ensuring terminology use is consistent and reflects the source text domain
		and by only ensuring you use equivalent idioms %s).\n\nWrite a list of specific, helpful and constructive suggestions for improving the translation. Each
		suggestion should address one specific part of the translation. Output only the suggestions and nothing else.`, sourceLang, targetLang, targetLang, country,
			sourceText, translation1, targetLang, targetLang)
	} else {
		reflectionPrompt = fmt.Sprintf(`Your task is to carefully read a source text and a translation from %s to %s, and then give constructive criticisms
		and helpful suggestions to improve the translation. \n\nThe source text and initial translation, delimited by XML tags <SOURCE_TEXT></SOURCE_TEXT> and
		<TRANSLATION></TRANSLATION>, are as follows:\n\n<SOURCE_TEXT>\n%s\n</SOURCE_TEXT>\n\n<TRANSLATION>\n%s\n</TRANSLATION>\n\nWhen writing suggestions, pay
		attention to whether there are ways to improve the translation's (i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated
		text), (ii) fluency (by applying %s grammar, spelling and punctuation rules, and ensuring there are no unnecessary repetitions), (iii) style (by ensuring th
		translations reflect the style of the source text and take into account any cultural context), (iv) terminology (by ensuring terminology use is consistent a
		reflects the source text domain; and by only ensuring you use equivalent idioms %s).\n\nWrite a list of specific, helpful and constructive suggestions for
			improving the translation. Each suggestion should address one specific part of the translation. Output only the suggestions and nothing else.`, sourceLang,
			targetLang, sourceText, translation1, targetLang, targetLang)
	}

	start := time.Now()
	defer func() {
		if err != nil {
			log.Errorf("oneChunkReflectOnTranslation error: %v", err)
			return
		}

		inputToken, _ := numTokensInString(reflectionPrompt, "cl100k_base")
		oneStep := &common.OneAiRecord{
			Cost:       time.Since(start).Milliseconds(),
			InputToken: inputToken,
		}
		oneStep.OutputToken, _ = numTokensInString(reflection, "cl100k_base")
		stepRecord.Records = append(stepRecord.Records, oneStep)
	}()

	reflection, err = t.ai.GetCompletion(reflectionPrompt, systemMessage, t.defaultModel, 0.3, false, ctx)
	if err != nil {
		return "", err
	}

	return reflection, nil
}

func (t *TranslateApp) oneChunkImproveTranslation(sourceLang, targetLang, sourceText, translation1, reflection string, stepRecord *common.TranslateStepRecord, ctx context.Context) (translation2 string, err error) {
	systemMessage := fmt.Sprintf("You are an expert linguist, specializing in translation editing from %s to %s.", sourceLang, targetLang)

	prompt := fmt.Sprintf(`Your task is to carefully read, then edit, a translation from %s to %s, taking into account a list of expert suggestions and
	constructive criticisms.\n\nThe source text, the initial translation, and the expert linguist suggestions are delimited by XML tags
	<SOURCE_TEXT></SOURCE_TEXT>, <TRANSLATION></TRANSLATION> and <EXPERT_SUGGESTIONS></EXPERT_SUGGESTIONS> as
follows:\n\n<SOURCE_TEXT>\n%s\n</SOURCE_TEXT>\n\n<TRANSLATION>\n%s\n</TRANSLATION>\n\n<EXPERT_SUGGESTIONS>\n%s\n</EXPERT_SUGGESTIONS>\n\nPlease take into
	account the expert suggestions when editing the translation. Edit the translation by ensuring:\n\n(i) accuracy (by correcting errors of addition,
		mistranslation, omission, or untranslated text), (ii) fluency (by applying %s grammar, spelling and punctuation rules and ensuring there are no unnecessary
	repetitions), (iii) style (by ensuring the translations reflect the style of the source text) (iv) terminology (inappropriate for context, inconsistent use)
	or (v) other errors.\n\nOutput only the new translation and nothing else.`, sourceLang, targetLang, sourceText, translation1, reflection, targetLang)

	start := time.Now()
	defer func() {
		if err != nil {
			log.Errorf("oneChunkImproveTranslation error: %v", err)
			return
		}

		inputToken, _ := numTokensInString(prompt, "cl100k_base")
		oneStep := &common.OneAiRecord{
			Cost:       time.Since(start).Milliseconds(),
			InputToken: inputToken,
		}
		oneStep.OutputToken, _ = numTokensInString(translation2, "cl100k_base")
		stepRecord.Records = append(stepRecord.Records, oneStep)
	}()

	translation2, err = t.ai.GetCompletion(prompt, systemMessage, t.defaultModel, 0.3, false, ctx)
	if err != nil {
		return "", err
	}

	return translation2, nil
}

func (t *TranslateApp) oneChunkTranslateText(sourceLang, targetLang, sourceText, country string, record *common.TranslateRecord, ctx context.Context) (string, error) {
	log.Infof("[single step1] start chunk initial translation for %s to %s", sourceLang, targetLang)
	translation1, err := func() (string, error) {
		start := time.Now()
		translation1 := ""
		var err error = nil

		step1 := &common.TranslateStepRecord{
			StepName: "simple_step_1",
			SumCost:  0,
			Records:  make([]*common.OneAiRecord, 0),
		}

		defer func() {
			if err != nil {
				return
			}
			step1.SumCost = time.Since(start).Milliseconds()
			record.Steps = append(record.Steps, step1)
		}()
		translation1, err = t.oneChunkInitialTranslation(sourceLang, targetLang, sourceText, step1, ctx)
		if err != nil {
			return "", err
		}

		return translation1, nil
	}()

	if err != nil {
		return "", err
	}

	log.Infof("[single step2] start chunk reflect on translation for %s to %s", sourceLang, targetLang)
	reflection, err := func() (string, error) {
		start := time.Now()
		reflection := ""
		var err error = nil
		step2 := &common.TranslateStepRecord{
			StepName: "simple_step_2",
			SumCost:  0,
			Records:  make([]*common.OneAiRecord, 0),
		}
		defer func() {
			if err != nil {
				return
			}
			step2.SumCost = time.Since(start).Milliseconds()
			record.Steps = append(record.Steps, step2)
		}()

		reflection, err = t.oneChunkReflectOnTranslation(sourceLang, targetLang, sourceText, translation1, country, step2, ctx)
		if err != nil {
			return "", err
		}
		return reflection, nil
	}()

	if err != nil {
		return "", err
	}

	log.Infof("[single step3] start chunk improve translation for %s to %s", sourceLang, targetLang)
	translation2, err := func() (string, error) {
		start := time.Now()
		translation2 := ""
		var err error = nil
		step3 := &common.TranslateStepRecord{
			StepName: "simple_step_3",
			SumCost:  0,
			Records:  make([]*common.OneAiRecord, 0),
		}
		defer func() {
			if err != nil {
				return
			}
			step3.SumCost = time.Since(start).Milliseconds()
			record.Steps = append(record.Steps, step3)
		}()
		translation2, err = t.oneChunkImproveTranslation(sourceLang, targetLang, sourceText, translation1, reflection, step3, ctx)
		if err != nil {
			return "", err
		}
		return translation2, nil
	}()
	if err != nil {
		return "", err
	}

	return translation2, nil
}

func numTokensInString(inputStr string, encodingName tokenizer.Encoding) (int, error) {
	encoding, err := tokenizer.Get(encodingName)
	if inputStr == "" {
		return 0, nil
	}

	if err != nil {
		log.Errorf("Failed to get encoding: %s", err)
		return 0, &common.InnerError{ErrType: common.ParameterError, ErrMsg: err.Error(), Code: 0}
	}

	tokens, _, err := encoding.Encode(inputStr)
	if err != nil {
		log.Errorf("Failed to encode: %s", err)
		return 0, &common.InnerError{ErrType: common.ExternalServiceError, ErrMsg: err.Error(), Code: 0}
	}
	return len(tokens), nil
}

func calculateChunkSize(tokenCount, tokenLimit int) int {
	if tokenCount <= tokenLimit {
		return tokenCount
	}

	numChunks := (tokenCount + tokenLimit - 1) / tokenLimit
	chunkSize := tokenCount / numChunks

	remainingTokens := tokenCount % tokenLimit
	if remainingTokens > 0 {
		chunkSize += remainingTokens / numChunks
	}

	return chunkSize
}

func (t *TranslateApp) Translate(sourceLang, targetLang, sourceText, country string, maxTokens int, ctx context.Context) (string, error) {
	if maxTokens <= 0 {
		maxTokens = t.maxToken
	}

	numTokensInText, err := numTokensInString(sourceText, "cl100k_base")
	if err != nil {
		return "", err
	}

	translateRecord := &common.TranslateRecord{
		MaxTokens:   maxTokens,
		InputTokens: numTokensInText,
		Steps:       make([]*common.TranslateStepRecord, 0),
		SourceLang:  sourceLang,
		TargetLang:  targetLang,
	}

	if numTokensInText <= maxTokens {
		log.Infof("Translating text as a single chunk, tokens:%d", numTokensInText)
		translateRecord.ChunksCount = 1
		finalTranslation, err := t.oneChunkTranslateText(sourceLang, targetLang, sourceText, country, translateRecord, ctx)
		if err != nil {
			return "", err
		}

		log.Infof("%s", translateRecord.String())
		return finalTranslation, nil
	} else {
		tokenSize := calculateChunkSize(numTokensInText, maxTokens)
		log.Infof("Translating text as multiple chunks, tokens:%d, chunk token:%d", numTokensInText, tokenSize)
		translateRecord.ChunkTokens = tokenSize

		textSplitter := textsplitter.NewRecursiveCharacter(textsplitter.WithChunkSize(tokenSize), textsplitter.WithChunkOverlap(0), textsplitter.WithModelName(t.defaultModel))
		sourceTextChunks, err := textSplitter.SplitText(sourceText)
		if err != nil {
			log.Errorf("Error splitting text: %v", err)
			return "", &common.InnerError{
				ErrType: common.ExternalServiceError,
				ErrMsg:  err.Error(),
				Code:    0,
			}
		}
		translateRecord.ChunksCount = len(sourceTextChunks)
		log.Infof("MultiTranslation chunks: %v", len(sourceTextChunks))

		translation2Chunks, err := t.MultiChunkTranslation(sourceLang, targetLang, sourceTextChunks, country, translateRecord, ctx)
		if err != nil {
			return "", &common.InnerError{
				ErrType: common.ExternalServiceError,
				ErrMsg:  err.Error(),
				Code:    0,
			}
		}

		log.Infof("%s", translateRecord.String())
		return strings.Join(translation2Chunks, ""), nil
	}
}

func (t *TranslateApp) multiChunkInitialTranslation(sourceLang, targetLang string, sourceTextChunks []string, record *common.TranslateStepRecord, ctx context.Context) ([]string, error) {
	systemMessage := fmt.Sprintf("You are an expert linguist, specializing in translation from %s to %s.", sourceLang, targetLang)
	translationPrompt := `Your task is to provide a professional translation from %s to %s of PART of a text.

 The source text is below, delimited by XML tags <SOURCE_TEXT> and </SOURCE_TEXT>. Translate only the part within the source text
 delimited by <TRANSLATE_THIS> and </TRANSLATE_THIS>. You can use the rest of the source text as context, but do not translate any
 of the other text. Do not output anything other than the translation of the indicated part of the text.

 <SOURCE_TEXT>
 %s
 </SOURCE_TEXT>

 To reiterate, you should translate only this part of the text, shown here again between <TRANSLATE_THIS> and </TRANSLATE_THIS>:
 <TRANSLATE_THIS>
 %s
 </TRANSLATE_THIS>

 Output only the translation of the portion you are asked to translate, and nothing else.`

	translationChunks := make([]string, 0)
	for i := range sourceTextChunks {
		translation, err := func(idx int) (string, error) {
			start := time.Now()
			startIdx, endIdx := t.getContextBoundary(i, len(sourceTextChunks))
			taggedText := strings.Join(sourceTextChunks[startIdx:i], "") + "<TRANSLATE_THIS>" + sourceTextChunks[i] + "</TRANSLATE_THIS>" +
				strings.Join(sourceTextChunks[i+1:endIdx], "")
			prompt := fmt.Sprintf(translationPrompt, sourceLang, targetLang, taggedText, sourceTextChunks[i])
			var err error = nil
			translation := ""

			defer func() {
				if err != nil {
					log.Errorf("[multi init chunk:%d]run voiceAi serive is error:%s", idx, err)
					return
				}
				inputToken, _ := numTokensInString(prompt, "cl100k_base")
				oneStep := &common.OneAiRecord{
					Cost:       time.Since(start).Milliseconds(),
					InputToken: inputToken,
				}
				oneStep.OutputToken, _ = numTokensInString(translation, "cl100k_base")
				record.Records = append(record.Records, oneStep)
			}()

			translation, err = t.ai.GetCompletion(prompt, systemMessage, t.defaultModel, 0.3, false, ctx)
			if err != nil {
				log.Errorf("run voiceAi service is error:%s", err)
				return "", &common.InnerError{
					ErrType: common.ExternalServiceError,
					ErrMsg:  err.Error(),
					Code:    0,
				}
			}
			return translation, nil
		}(i)

		if err != nil {
			return nil, err
		}
		translationChunks = append(translationChunks, translation)
	}
	return translationChunks, nil
}

func (t *TranslateApp) multiChunkReflectOnTranslation(sourceLang, targetLang string, sourceTextChunks, translation1Chunks []string, country string, record *common.TranslateStepRecord, ctx context.Context) ([]string, error) {
	systemMessage := fmt.Sprintf("You are an expert linguist specializing in translation from %s to %s. You will be provided with a source text and its translation and your goal is to improve the translation.", sourceLang, targetLang)

	var reflectionPrompt string
	if country != "" {
		reflectionPrompt = `Your task is to carefully read a source text and part of a translation of that text from %s to %s, and then give constructive
 criticism and helpful suggestions for improving the translation.
 The final style and tone of the translation should match the style of %s colloquially spoken in %s.

 The source text is below, delimited by XML tags <SOURCE_TEXT> and </SOURCE_TEXT>, and the part that has been translated
 is delimited by <TRANSLATE_THIS> and </TRANSLATE_THIS> within the source text. You can use the rest of the source text
 as context for critiquing the translated part.

 <SOURCE_TEXT>
 %s
 </SOURCE_TEXT>

 To reiterate, only part of the text is being translated, shown here again between <TRANSLATE_THIS> and </TRANSLATE_THIS>:
 <TRANSLATE_THIS>
 %s
 </TRANSLATE_THIS>

 The translation of the indicated part, delimited below by <TRANSLATION> and </TRANSLATION>, is as follows:
 <TRANSLATION>
 %s
 </TRANSLATION>

 When writing suggestions, pay attention to whether there are ways to improve the translation's:
 (i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),
 (ii) fluency (by applying %s grammar, spelling and punctuation rules, and ensuring there are no unnecessary repetitions),
 (iii) style (by ensuring the translations reflect the style of the source text and take into account any cultural context),
 (iv) terminology (by ensuring terminology use is consistent and reflects the source text domain; and by only ensuring you use equivalent idioms %s).

 Write a list of specific, helpful and constructive suggestions for improving the translation.
 Each suggestion should address one specific part of the translation.
 Output only the suggestions and nothing else.`
	} else {
		reflectionPrompt = `Your task is to carefully read a source text and part of a translation of that text from %s to %s, and then give constructive
 criticism and helpful suggestions for improving the translation.

 The source text is below, delimited by XML tags <SOURCE_TEXT> and </SOURCE_TEXT>, and the part that has been translated
 is delimited by <TRANSLATE_THIS> and </TRANSLATE_THIS> within the source text. You can use the rest of the source text
 as context for critiquing the translated part.

 <SOURCE_TEXT>
 %s
 </SOURCE_TEXT>

 To reiterate, only part of the text is being translated, shown here again between <TRANSLATE_THIS> and </TRANSLATE_THIS>:
 <TRANSLATE_THIS>
 %s
 </TRANSLATE_THIS>

 The translation of the indicated part, delimited below by <TRANSLATION> and </TRANSLATION>, is as follows:
 <TRANSLATION>
 %s
 </TRANSLATION>

 When writing suggestions, pay attention to whether there are ways to improve the translation's:
 (i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),
 (ii) fluency (by applying %s grammar, spelling and punctuation rules, and ensuring there are no unnecessary repetitions),
 (iii) style (by ensuring the translations reflect the style of the source text and take into account any cultural context),
 (iv) terminology (by ensuring terminology use is consistent and reflects the source text domain; and by only ensuring you use equivalent idioms %s).

 Write a list of specific, helpful and constructive suggestions for improving the translation.
 Each suggestion should address one specific part of the translation.
 Output only the suggestions and nothing else.`
	}

	reflectionChunks := make([]string, 0)
	for i := range sourceTextChunks {
		reflection, err := func(idx int) (string, error) {
			startIdx, endIdx := t.getContextBoundary(i, len(sourceTextChunks))
			taggedText := strings.Join(sourceTextChunks[startIdx:i], "") + "<TRANSLATE_THIS>" + sourceTextChunks[i] + "</TRANSLATE_THIS>" +
				strings.Join(sourceTextChunks[i+1:endIdx], "")
			reflection := ""
			prompt := fmt.Sprintf(reflectionPrompt, sourceLang, targetLang, taggedText, sourceTextChunks[i], translation1Chunks[i], targetLang, targetLang)
			var err error = nil
			if country != "" {
				prompt = fmt.Sprintf(reflectionPrompt, sourceLang, targetLang, targetLang, country, taggedText, sourceTextChunks[i], translation1Chunks[i],
					targetLang, targetLang)
			}
			start := time.Now()
			defer func() {
				if err != nil {
					log.Errorf("[multi refection chunk:%d]run voiceAi serive is error:%s", idx, err)
					return
				}
				inputToken, _ := numTokensInString(prompt, "cl100k_base")
				oneStep := &common.OneAiRecord{
					Cost:       time.Since(start).Milliseconds(),
					InputToken: inputToken,
				}
				oneStep.OutputToken, _ = numTokensInString(reflection, "cl100k_base")
				record.Records = append(record.Records, oneStep)
			}()

			reflection, err = t.ai.GetCompletion(prompt, systemMessage, t.defaultModel, 0.3, false, ctx)
			return reflection, err
		}(i)
		if err != nil {
			return nil, err
		}
		reflectionChunks = append(reflectionChunks, reflection)
	}

	return reflectionChunks, nil
}

func (t *TranslateApp) multiChunkImproveTranslation(sourceLang, targetLang string, sourceTextChunks, translation1Chunks, reflectionChunks []string, record *common.TranslateStepRecord, ctx context.Context) ([]string, error) {
	systemMessage := fmt.Sprintf("You are an expert linguist, specializing in translation editing from %s to %s.", sourceLang, targetLang)

	improvementPrompt := `Your task is to carefully read, then improve, a translation from %s to %s, taking into
 account a set of expert suggestions and constructive criticisms. Below, the source text, initial translation, and expert suggestions are provided.

 The source text is below, delimited by XML tags <SOURCE_TEXT> and </SOURCE_TEXT>, and the part that has been translated
 is delimited by <TRANSLATE_THIS> and </TRANSLATE_THIS> within the source text. You can use the rest of the source text
 as context, but need to provide a translation only of the part indicated by <TRANSLATE_THIS> and </TRANSLATE_THIS>.

 <SOURCE_TEXT>
 %s
 </SOURCE_TEXT>

 To reiterate, only part of the text is being translated, shown here again between <TRANSLATE_THIS> and </TRANSLATE_THIS>:
 <TRANSLATE_THIS>
 %s
 </TRANSLATE_THIS>

 The translation of the indicated part, delimited below by <TRANSLATION> and </TRANSLATION>, is as follows:
 <TRANSLATION>
 %s
 </TRANSLATION>

 The expert translations of the indicated part, delimited below by <EXPERT_SUGGESTIONS> and </EXPERT_SUGGESTIONS>, are as follows:
 <EXPERT_SUGGESTIONS>
 %s
 </EXPERT_SUGGESTIONS>

 Taking into account the expert suggestions rewrite the translation to improve it, paying attention
 to whether there are ways to improve the translation's

 (i) accuracy (by correcting errors of addition, mistranslation, omission, or untranslated text),
 (ii) fluency (by applying %s grammar, spelling and punctuation rules and ensuring there are no unnecessary repetitions),
 (iii) style (by ensuring the translations reflect the style of the source text)
 (iv) terminology (inappropriate for context, inconsistent use), or
 (v) other errors.

 Output only the new translation of the indicated part and nothing else.`

	translation2Chunks := make([]string, 0)
	for i := range sourceTextChunks {
		translation2, err := func(idx int) (string, error) {
			startIdx, endIdx := t.getContextBoundary(i, len(sourceTextChunks))
			taggedText := strings.Join(sourceTextChunks[startIdx:i], "") + "<TRANSLATE_THIS>" + sourceTextChunks[i] + "</TRANSLATE_THIS>" +
				strings.Join(sourceTextChunks[i+1:endIdx], "")
			prompt := fmt.Sprintf(improvementPrompt, sourceLang, targetLang, taggedText, sourceTextChunks[i], translation1Chunks[i], reflectionChunks[i],
				targetLang)
			translation2 := ""
			var err error = nil
			start := time.Now()

			defer func() {
				if err != nil {
					log.Errorf("[multi improvement chunk:%d]run voiceAi serive is error:%s", idx, err)
					return
				}
				inputToken, _ := numTokensInString(prompt, "cl100k_base")
				oneStep := &common.OneAiRecord{
					Cost:       time.Since(start).Milliseconds(),
					InputToken: inputToken,
				}
				oneStep.OutputToken, _ = numTokensInString(translation2, "cl100k_base")
				record.Records = append(record.Records, oneStep)
			}()
			translation2, err = t.ai.GetCompletion(prompt, systemMessage, t.defaultModel, 0.3, false, ctx)

			if err != nil {
				return "", err
			}
			return translation2, nil
		}(i)

		if err != nil {
			return nil, err
		}
		translation2Chunks = append(translation2Chunks, translation2)
	}

	return translation2Chunks, nil
}

func (t *TranslateApp) MultiChunkTranslation(sourceLang, targetLang string, sourceTextChunks []string, country string, record *common.TranslateRecord, ctx context.Context) ([]string, error) {

	translation1Chunks, err := func() ([]string, error) {
		translation1Chunks := make([]string, 0)
		var err error = nil
		start := time.Now()

		step1 := &common.TranslateStepRecord{
			StepName: "multi_step_1",
			SumCost:  0,
			Records:  []*common.OneAiRecord{},
		}

		defer func() {
			if err != nil {
				return
			}
			step1.SumCost = time.Since(start).Milliseconds()
			record.Steps = append(record.Steps, step1)
		}()

		translation1Chunks, err = t.multiChunkInitialTranslation(sourceLang, targetLang, sourceTextChunks, step1, ctx)
		return translation1Chunks, err
	}()
	if err != nil {
		return nil, err
	}

	reflectionChunks, err := func() ([]string, error) {
		reflectionChunks := make([]string, 0)
		var err error = nil
		start := time.Now()

		step2 := &common.TranslateStepRecord{
			StepName: "multi_step_2",
			SumCost:  0,
			Records:  []*common.OneAiRecord{},
		}
		defer func() {
			if err != nil {
				return
			}
			step2.SumCost = time.Since(start).Milliseconds()
			record.Steps = append(record.Steps, step2)
		}()
		reflectionChunks, err = t.multiChunkReflectOnTranslation(sourceLang, targetLang, sourceTextChunks, translation1Chunks, country, step2, ctx)
		return reflectionChunks, err
	}()
	if err != nil {
		return nil, err
	}

	translation2Chunks, err := func() ([]string, error) {
		translation2Chunks := make([]string, 0)
		var err error = nil
		start := time.Now()

		step3 := &common.TranslateStepRecord{
			StepName: "multi_step_3",
			SumCost:  0,
			Records:  []*common.OneAiRecord{},
		}
		defer func() {
			if err != nil {
				return
			}
			step3.SumCost = time.Since(start).Milliseconds()
			record.Steps = append(record.Steps, step3)
		}()
		translation2Chunks, err = t.multiChunkImproveTranslation(sourceLang, targetLang, sourceTextChunks, translation1Chunks, reflectionChunks, step3, ctx)
		return translation2Chunks, err
	}()
	if err != nil {
		return nil, err
	}

	return translation2Chunks, nil
}

func (t *TranslateApp) getContextBoundary(i int, maxBoundary int) (int, int) {
	offset := t.chunksInContext / 2

	startIdx := i - offset
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := i + offset + 1
	if endIdx >= maxBoundary {
		endIdx = maxBoundary
	}
	return startIdx, endIdx
}
