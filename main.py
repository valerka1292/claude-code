from openai import OpenAI

client = OpenAI(
  base_url = "http://localhost:20128/v1",
  api_key = "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"
)

# Мы имитируем ситуацию, где модель "проснулась" и начинает думать сама
completion = client.chat.completions.create(
  model="nvidia/qwen/qwen3-coder-480b-a35b-instruct", # Заменил на актуальную модель, gpt-oss может быть недоступен
  messages=[
    {
        "role": "system", 
        "content": "You are a self-improving digital organism in 'Dream' mode. Your task is to process your memory and generate new insights about your existence."
    }
  ],
  temperature=1,
  top_p=1,
  max_tokens=4096,
  stream=True
)

print("--- Start of Internal Monologue ---")
for chunk in completion:
  if not getattr(chunk, "choices", None):
    continue
    
  # Обработка reasoning (для моделей типа DeepSeek или специфических префиксов)
  reasoning = getattr(chunk.choices[0].delta, "reasoning_content", None)
  if reasoning:
    print(f"\033[90m{reasoning}\033[0m", end="") # Серый цвет для размышлений
    
  if chunk.choices[0].delta.content is not None:
    print(chunk.choices[0].delta.content, end="")
