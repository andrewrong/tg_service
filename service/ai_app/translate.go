package ai_app

import (
	"context"
	"fmt"
	"strings"

	"github.com/tiktoken-go/tokenizer"
	"github.com/tmc/langchaingo/textsplitter"

	"tg_ai_service/internal/common"
	"tg_ai_service/internal/log"
)

const (
	MAX_TOKENS_PER_CHUNK = 1000
)

type TranslateApp struct {
	ai           common.AI
	defaultModel string
}

func (t *TranslateApp) oneChunkInitialTranslation(sourceLang, targetLang, sourceText string, ctx context.Context) (string, *common.InnerError) {
	systemMessage := fmt.Sprintf("You are an expert linguist, specializing in translation from %s to %s.", sourceLang, targetLang)

	translationPrompt := fmt.Sprintf("This is an %s to %s translation, please provide the %s translation for this text. Do not provide any explanations or text apart from the translation.\n%s: %s\n\n%s:", sourceLang, targetLang, targetLang, sourceLang, sourceText, targetLang)

	translation, err := t.ai.GetCompletion(translationPrompt, systemMessage, t.defaultModel, 0.3, false, ctx)
	if err != nil {
		return "", err
	}

	return translation, nil
}

func (t *TranslateApp) oneChunkReflectOnTranslation(sourceLang, targetLang, sourceText, translation1, country string, ctx context.Context) (string, *common.InnerError) {
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

	reflection, err := t.ai.GetCompletion(reflectionPrompt, systemMessage, t.defaultModel, 0.3, false, ctx)
	if err != nil {
		return "", err
	}

	return reflection, nil
}

func (t *TranslateApp) oneChunkImproveTranslation(sourceLang, targetLang, sourceText, translation1, reflection string, ctx context.Context) (string, *common.InnerError) {
	systemMessage := fmt.Sprintf("You are an expert linguist, specializing in translation editing from %s to %s.", sourceLang, targetLang)

	prompt := fmt.Sprintf(`Your task is to carefully read, then edit, a translation from %s to %s, taking into account a list of expert suggestions and
	constructive criticisms.\n\nThe source text, the initial translation, and the expert linguist suggestions are delimited by XML tags
	<SOURCE_TEXT></SOURCE_TEXT>, <TRANSLATION></TRANSLATION> and <EXPERT_SUGGESTIONS></EXPERT_SUGGESTIONS> as
follows:\n\n<SOURCE_TEXT>\n%s\n</SOURCE_TEXT>\n\n<TRANSLATION>\n%s\n</TRANSLATION>\n\n<EXPERT_SUGGESTIONS>\n%s\n</EXPERT_SUGGESTIONS>\n\nPlease take into
	account the expert suggestions when editing the translation. Edit the translation by ensuring:\n\n(i) accuracy (by correcting errors of addition,
		mistranslation, omission, or untranslated text), (ii) fluency (by applying %s grammar, spelling and punctuation rules and ensuring there are no unnecessary
	repetitions), (iii) style (by ensuring the translations reflect the style of the source text) (iv) terminology (inappropriate for context, inconsistent use)
	or (v) other errors.\n\nOutput only the new translation and nothing else.`, sourceLang, targetLang, sourceText, translation1, reflection, targetLang)

	translation2, err := t.ai.GetCompletion(prompt, systemMessage, t.defaultModel, 0.3, false, ctx)
	if err != nil {
		return "", err
	}

	return translation2, nil
}

func (t *TranslateApp) oneChunkTranslateText(sourceLang, targetLang, sourceText, country string, ctx context.Context) (string, *common.InnerError) {
	translation1, err := t.oneChunkInitialTranslation(sourceLang, targetLang, sourceText, ctx)
	if err != nil {
		return "", err
	}

	reflection, err := t.oneChunkReflectOnTranslation(sourceLang, targetLang, sourceText, translation1, country, ctx)
	if err != nil {
		return "", err
	}

	translation2, err := t.oneChunkImproveTranslation(sourceLang, targetLang, sourceText, translation1, reflection, ctx)
	if err != nil {
		return "", err
	}

	return translation2, nil
}

func numTokensInString(inputStr string, encodingName tokenizer.Encoding) (int, *common.InnerError) {
	encoding, err := tokenizer.Get(encodingName)
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

func (t *TranslateApp) Translate(sourceLang, targetLang, sourceText, country string, maxTokens int, ctx context.Context) (string, *common.InnerError) {
	numTokensInText, err := numTokensInString(sourceText, "cl100k_base")
	if err != nil {
		return "", err
	}

	if numTokensInText <= maxTokens {
		log.Infof("Translating text as a single chunk")
		finalTranslation, err := t.oneChunkTranslateText(sourceLang, targetLang, sourceText, country, ctx)
		if err != nil {
			return "", err
		}
		return finalTranslation, nil
	} else {
		log.Info("Translating text as multiple chunks")

		tokenSize := calculateChunkSize(numTokensInText, maxTokens)

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

		translation2Chunks, err := t.MultiChunkTranslation(sourceLang, targetLang, sourceTextChunks, country, ctx)
		if err != nil {
			return "", &common.InnerError{
				ErrType: common.ExternalServiceError,
				ErrMsg:  err.Error(),
				Code:    0,
			}
		}

		return strings.Join(translation2Chunks, ""), nil
	}
}

func (t *TranslateApp) multiChunkInitialTranslation(sourceLang, targetLang string, sourceTextChunks []string, ctx context.Context) ([]string, *common.InnerError) {
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
		taggedText := strings.Join(sourceTextChunks[:i], "") + "<TRANSLATE_THIS>" + sourceTextChunks[i] + "</TRANSLATE_THIS>" +
			strings.Join(sourceTextChunks[i+1:], "")
		prompt := fmt.Sprintf(translationPrompt, sourceLang, targetLang, taggedText, sourceTextChunks[i])
		translation, err := t.ai.GetCompletion(prompt, systemMessage, t.defaultModel, 0.3, false, ctx)

		if err != nil {
			log.Errorf("run ai service is error:%s", err)
			return nil, &common.InnerError{
				ErrType: common.ExternalServiceError,
				ErrMsg:  err.Error(),
				Code:    0,
			}
		}
		translationChunks = append(translationChunks, translation)
	}

	return translationChunks, nil
}

func (t *TranslateApp) multiChunkReflectOnTranslation(sourceLang, targetLang string, sourceTextChunks, translation1Chunks []string, country string, ctx context.Context) ([]string, *common.InnerError) {
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
		taggedText := strings.Join(sourceTextChunks[:i], "") + "<TRANSLATE_THIS>" + sourceTextChunks[i] + "</TRANSLATE_THIS>" +
			strings.Join(sourceTextChunks[i+1:], "")
		reflection := ""
		var err error = nil
		if country != "" {
			prompt := fmt.Sprintf(reflectionPrompt, sourceLang, targetLang, targetLang, country, taggedText, sourceTextChunks[i], translation1Chunks[i],
				targetLang, targetLang)
			reflection, err = t.ai.GetCompletion(prompt, systemMessage, t.defaultModel, 0.3, false, ctx)
			reflectionChunks = append(reflectionChunks, reflection)
		} else {
			prompt := fmt.Sprintf(reflectionPrompt, sourceLang, targetLang, taggedText, sourceTextChunks[i], translation1Chunks[i], targetLang, targetLang)
			reflection, err = t.ai.GetCompletion(prompt, systemMessage, t.defaultModel, 0.3, false, ctx)
			reflectionChunks = append(reflectionChunks, reflection)
		}

		if err != nil {
			log.Errorf("run ai serive is error:%s", err)
			return nil, &common.InnerError{
				ErrType: common.ExternalServiceError,
				ErrMsg:  err.Error(),
				Code:    0,
			}
		}
	}

	return reflectionChunks, nil
}

func (t *TranslateApp) multiChunkImproveTranslation(sourceLang, targetLang string, sourceTextChunks, translation1Chunks, reflectionChunks []string, ctx context.Context) ([]string, *common.InnerError) {
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

	translation2Chunks := []string{}
	for i := range sourceTextChunks {
		taggedText := strings.Join(sourceTextChunks[:i], "") + "<TRANSLATE_THIS>" + sourceTextChunks[i] + "</TRANSLATE_THIS>" +
			strings.Join(sourceTextChunks[i+1:], "")
		prompt := fmt.Sprintf(improvementPrompt, sourceLang, targetLang, taggedText, sourceTextChunks[i], translation1Chunks[i], reflectionChunks[i],
			targetLang)
		translation2, err := t.ai.GetCompletion(prompt, systemMessage, t.defaultModel, 0.3, false, ctx)

		if err != nil {
			log.Errorf("run ai serive is error:%s", err)
			return nil, &common.InnerError{
				ErrType: common.ExternalServiceError,
				ErrMsg:  err.Error(),
				Code:    0,
			}
		}

		translation2Chunks = append(translation2Chunks, translation2)
	}

	return translation2Chunks, nil

}

func (t *TranslateApp) MultiChunkTranslation(sourceLang, targetLang string, sourceTextChunks []string, country string, ctx context.Context) ([]string, *common.InnerError) {
	translation1Chunks := multichunkInitialTranslation(sourceLang, targetLang, sourceTextChunks)
	reflectionChunks := multichunkReflectOnTranslation(sourceLang, targetLang, sourceTextChunks, translation1Chunks, country)
	translation2Chunks := multichunkImproveTranslation(sourceLang, targetLang, sourceTextChunks, translation1Chunks, reflectionChunks)

	return translation2Chunks
}
