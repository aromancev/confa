import { computed, ComputedRef, Ref } from "vue"

export class RegexValidator {
  regex: RegExp
  error: string

  constructor(regex: string, errors: string[]) {
    this.regex = new RegExp(regex)
    this.error = errors.map((err: string) => "• " + err).join("\n")
  }

  validate(value: string): string {
    if (this.regex.test(value)) {
      return ""
    }
    return this.error
  }

  reactive(ref: Ref<string>): ComputedRef<string> {
    return computed(() => {
      return this.validate(ref.value)
    })
  }
}
