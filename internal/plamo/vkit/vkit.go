package vkit

import (
	"github.com/ensoria/rest/pkg/rest"
	"github.com/ensoria/validator/pkg/rule"
	"github.com/ensoria/validator/pkg/validate"
	"github.com/ensoria/validator/pkg/verr"
)

// validate.HttpRequestBodyFieldFirst か、
// validate.HttpRequestBodyLangFirst はプロジェクトの都合に合わせて
// 選択してください。
func RestRequestBody[T any](r *rest.Request, ruleSets ...*rule.RuleSet) (*T, verr.ValidationErrorMessages) {
	return validate.RestRequestBodyFieldFirst[T](r, ruleSets...)
}

// validate.MapFieldFirst か、
// validate.MapLangFirst はプロジェクトの都合に合わせて
// 選択してください。
func Map[T any](m map[string]T, ruleSets ...*rule.RuleSet) verr.ValidationErrorMessages {
	return validate.MapFieldFirst(m, ruleSets...)
}

// 以下のバリデーションは、共通で使えるようなメッセージで定義しています。
// より詳細なメッセージでメッセージを定義したい場合は、各Module内で
// 別のメッセージのRuleFactoryの定義を作成してください。

var MaxLength = rule.CreateStrMaxLength(
	map[string]string{
		"ja": "最大文字数%dを超えています",
		"en": "exceeds maximum length of %d characters",
	})

// TODO: 他のバリデーションも一通り定義する
