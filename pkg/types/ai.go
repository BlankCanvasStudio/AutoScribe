package types;

type Model string;

const (
    NoModel     Model = ""
    GPT_41_Nano Model = "gpt 4.1 nano"
)

var Models []Model = []Model {
    NoModel,
    GPT_41_Nano,
}
