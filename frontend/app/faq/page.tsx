"use client";

import { useState } from "react";
import Link from "next/link";
import { ChevronDown, ArrowLeft, HelpCircle, BookOpen, Sun, Moon } from "lucide-react";
import { useTheme } from "@/lib/theme";

type Lang = "en" | "ru" | "kk";

interface FAQItem {
  q: string;
  a: string;
  html?: boolean;
}

const FAQ_DATA: Record<Lang, FAQItem[]> = {
  ru: [
    {
      q: "Сроки подачи заявок на осенний набор 2026?",
      a: `<strong>Ранний прием:</strong> до 24 декабря 2025 года<br/><br/>Собеседование: проводится с представителем университета в конце декабря 2025 — начале января 2026 года. Пожалуйста, регулярно проверяйте электронную почту на наличие письма от Приёмной комиссии о назначении даты собеседования.<br/><br/><strong>Обычный прием:</strong> весна 2026 года. Обратите внимание, что наличие свободных мест на данном этапе не гарантируется, поэтому настоятельно рекомендуем подать заявку как можно раньше.`,
      html: true,
    },
    {
      q: "Нужно ли сдавать ЕНТ?",
      a: "Да, для граждан Казахстана это обязательно. Так как срок раннего приёма заканчивается в декабре 2025 года, настоятельно рекомендуется сдавать ЕНТ в январе 2026 года.",
    },
    {
      q: "Какие тесты по английскому принимаются?",
      a: "IELTS, TOEFL или Duolingo (для иностранных кандидатов, у которых нет возможности сдать IELTS или TOEFL).",
    },
    {
      q: "Обязательно ли проходить собеседование?",
      a: "Это обязательная часть приёма. Вам зададут вопросы о вашей мотивации и целях. Используйте эту возможность, чтобы объяснить, почему вы являетесь хорошим кандидатом для программы и университета.",
    },
    {
      q: "Я забыл загрузить документы, что делать?",
      a: "Вы можете дозагрузить их через портал. Файлы, не загруженные вовремя, могут быть рассмотрены позже.",
    },
    {
      q: "Можно ли подать заявку из других стран Центральной Азии?",
      a: "Да, программа открыта для выпускников Казахстана, Кыргызстана, Узбекистана, Таджикистана и Туркменистана.",
    },
    {
      q: "Какой язык обучения?",
      a: "Основной язык — английский.",
    },
    {
      q: "Есть ли финансовая помощь?",
      a: "Да, все студенты обучаются на гранте.",
    },
    {
      q: "Сколько студентов принимается?",
      a: "Мы планируем набирать 100 студентов ежегодно на пять программ.",
    },
    {
      q: "Можно ли посетить кампус?",
      a: 'Да, посещения планируются к осени 2026, следите за обновлениями в <a href="https://www.instagram.com/invisionu.indrive/" target="_blank" rel="noopener noreferrer" class="text-primary underline">Instagram inVision U</a>.',
      html: true,
    },
  ],
  en: [
    {
      q: "What are the application deadlines for the Fall 2026 cohort?",
      a: `<strong>Early admission:</strong> until December 24, 2025.<br/><br/>Interviews are conducted with a university representative in late December 2025 – early January 2026. Please check your email regularly for a message from the Admissions Committee regarding your interview date.<br/><br/><strong>Regular admission:</strong> Spring 2026. Note that seat availability at this stage is not guaranteed, so we strongly recommend applying as early as possible.`,
      html: true,
    },
    {
      q: "Is the ENT (National Exam) required?",
      a: "Yes, it is mandatory for citizens of Kazakhstan. Since the early admission deadline is in December 2025, we strongly recommend taking the ENT in January 2026.",
    },
    {
      q: "Which English proficiency tests are accepted?",
      a: "IELTS, TOEFL, or Duolingo (for international applicants who are unable to take IELTS or TOEFL).",
    },
    {
      q: "Is the interview mandatory?",
      a: "Yes, it is a required part of the admissions process. You will be asked questions about your motivation and goals. Use this opportunity to explain why you are a strong candidate for the program and university.",
    },
    {
      q: "I forgot to upload documents — what should I do?",
      a: "You can upload them later through the portal. Files not submitted on time may be reviewed at a later stage.",
    },
    {
      q: "Can I apply from other Central Asian countries?",
      a: "Yes, the program is open to graduates from Kazakhstan, Kyrgyzstan, Uzbekistan, Tajikistan, and Turkmenistan.",
    },
    {
      q: "What is the language of instruction?",
      a: "The primary language of instruction is English.",
    },
    {
      q: "Is there financial support?",
      a: "Yes, all students study on a full grant (scholarship).",
    },
    {
      q: "How many students will be admitted?",
      a: "We plan to admit 100 students annually across five programs.",
    },
    {
      q: "Can I visit the campus?",
      a: 'Yes, campus visits are planned for autumn 2026. Follow updates on <a href="https://www.instagram.com/invisionu.indrive/" target="_blank" rel="noopener noreferrer" class="text-primary underline">Instagram inVision U</a>.',
      html: true,
    },
  ],
  kk: [
    {
      q: "2026 күзгі қабылдауға өтінім берудің мерзімдері?",
      a: `<strong>Ерте қабылдау:</strong> 2025 жылдың 24 желтоқсанына дейін.<br/><br/>Сұхбат 2025 жылдың желтоқсан аяғы — 2026 жылдың қаңтар басында университет өкілімен өткізіледі. Сұхбат күнін тағайындау туралы Қабылдау комиссиясының хабарламасы үшін электрондық поштаңызды жиі тексеріңіз.<br/><br/><strong>Жалпы қабылдау:</strong> 2026 жылдың көктемі. Бұл кезеңде бос орындар кепілдік берілмейді, сондықтан мүмкіндігінше ерте өтінім беруді ұсынамыз.`,
      html: true,
    },
    {
      q: "ҰБТ тапсыру міндетті ме?",
      a: "Иә, Қазақстан азаматтары үшін міндетті. Ерте қабылдаудың мерзімі 2025 жылдың желтоқсанында аяқталатындықтан, ҰБТ-ны 2026 жылдың қаңтарында тапсыруды ұсынамыз.",
    },
    {
      q: "Қандай ағылшын тілі сертификаттары қабылданады?",
      a: "IELTS, TOEFL немесе Duolingo (IELTS немесе TOEFL тапсыру мүмкіндігі жоқ шетелдік үміткерлер үшін).",
    },
    {
      q: "Сұхбат міндетті ме?",
      a: "Иә, бұл қабылдаудың міндетті бөлігі. Сізге мотивация мен мақсаттар туралы сұрақтар қойылады. Неліктен бағдарламаға жақсы үміткер екеніңізді түсіндіру мүмкіндігін пайдаланыңыз.",
    },
    {
      q: "Құжаттарды жүктеуді ұмытып қалдым, не істеу керек?",
      a: "Оларды портал арқылы кейінірек жүктей аласыз. Уақытында жүктелмеген файлдар кейінірек қаралуы мүмкін.",
    },
    {
      q: "Орталық Азияның басқа елдерінен өтінім беруге бола ма?",
      a: "Иә, бағдарлама Қазақстан, Қырғызстан, Өзбекстан, Тәжікстан және Түрікменстан түлектеріне ашық.",
    },
    {
      q: "Оқыту тілі қандай?",
      a: "Негізгі оқыту тілі — ағылшын.",
    },
    {
      q: "Қаржылық көмек бар ма?",
      a: "Иә, барлық студенттер толық грантпен оқиды.",
    },
    {
      q: "Қанша студент қабылданады?",
      a: "Біз бес бағдарлама бойынша жыл сайын 100 студент қабылдауды жоспарлаймыз.",
    },
    {
      q: "Кампусқа бару мүмкін бе?",
      a: 'Иә, кампусқа бару 2026 жылдың күзіне жоспарланған. Жаңартулар үшін <a href="https://www.instagram.com/invisionu.indrive/" target="_blank" rel="noopener noreferrer" class="text-primary underline">Instagram inVision U</a> бетін қадағалаңыз.',
      html: true,
    },
  ],
};

const REQUIREMENTS: Record<Lang, { heading: string; items: { number: string; title: string; content: string }[] }> = {
  ru: {
    heading: "Минимальные требования",
    items: [
      {
        number: "01",
        title: "ЕНТ",
        content: "80 баллов (для граждан Казахстана) по одной из связок:\n• Математика + География → Социология инноваций и лидерства, Стратегии государственного управления\n• Математика + Информатика → Инновационные цифровые продукты\n• Математика + Физика → Креативная инженерия\n• История Казахстана + Грамотность чтения + 2 творческих экзамена → Цифровые медиа",
      },
      {
        number: "02",
        title: "Уровень английского",
        content: "IELTS 6.0 / TOEFL iBT 60–78 / Duolingo 105–115",
      },
      {
        number: "03",
        title: "Документы",
        content: "• Удостоверение личности или паспорт\n• Ссылка на видеопрезентацию\n• Сертификат ЕНТ (для граждан Казахстана)\n• Сертификат уровня английского",
      },
    ],
  },
  en: {
    heading: "Basic Requirements",
    items: [
      {
        number: "01",
        title: "ENT Score",
        content: "80 points (for citizens of Kazakhstan) in one of the following combinations:\n• Math + Geography → Sociology: Leadership and Innovation, Public Policy\n• Math + Informatics → Innovative IT Product Design\n• Math + Physics → Creative Engineering\n• History of Kazakhstan + Reading Literacy + 2 creative exams → Digital Media",
      },
      {
        number: "02",
        title: "English Proficiency",
        content: "IELTS 6.0 / TOEFL iBT 60–78 / Duolingo 105–115",
      },
      {
        number: "03",
        title: "Required Documents",
        content: "• ID card or passport\n• Link to video presentation\n• ENT certificate (for citizens of Kazakhstan)\n• English proficiency certificate",
      },
    ],
  },
  kk: {
    heading: "Минималды талаптар",
    items: [
      {
        number: "01",
        title: "ҰБТ",
        content: "80 балл (Қазақстан азаматтары үшін) мына пәндер жиынтығының бірі бойынша:\n• Математика + География → Инновация және көшбасшылық социологиясы, Мемлекеттік басқару\n• Математика + Информатика → Инновациялық цифрлық өнімдер\n• Математика + Физика → Креативті инженерия\n• Қазақстан тарихы + Оқу сауаттылығы + 2 шығармашылық емтихан → Цифрлық медиа",
      },
      {
        number: "02",
        title: "Ағылшын тілі деңгейі",
        content: "IELTS 6.0 / TOEFL iBT 60–78 / Duolingo 105–115",
      },
      {
        number: "03",
        title: "Қажетті құжаттар",
        content: "• Жеке куәлік немесе паспорт\n• Бейне-таныстырылымға сілтеме\n• ҰБТ сертификаты (Қазақстан азаматтары үшін)\n• Ағылшын тілін меңгеру сертификаты",
      },
    ],
  },
};

const MAJORS_INFO: Record<Lang, { tag: string; name: string }[]> = {
  en: [
    { tag: "Engineering", name: "Creative Engineering" },
    { tag: "Tech", name: "Innovative IT Product Design and Development" },
    { tag: "Society", name: "Sociology: Leadership and Innovation" },
    { tag: "Policy Reform", name: "Public Policy and Development" },
    { tag: "Art + Media", name: "Digital Media and Marketing" },
  ],
  ru: [
    { tag: "Engineering", name: "Креативная инженерия" },
    { tag: "Tech", name: "Инновационные цифровые продукты и сервисы" },
    { tag: "Society", name: "Социология инноваций и лидерства" },
    { tag: "Policy Reform", name: "Стратегии государственного управления и развития" },
    { tag: "Art + Media", name: "Цифровые медиа и маркетинг" },
  ],
  kk: [
    { tag: "Engineering", name: "Креативті инженерия" },
    { tag: "Tech", name: "Инновациялық цифрлық өнімдер мен қызметтер" },
    { tag: "Society", name: "Инновация және көшбасшылық социологиясы" },
    { tag: "Policy Reform", name: "Мемлекеттік басқару және даму стратегиялары" },
    { tag: "Art + Media", name: "Цифрлық медиа және маркетинг" },
  ],
};

const UI_LABELS: Record<Lang, Record<string, string>> = {
  en: {
    "faq.title": "Frequently Asked Questions",
    "faq.subtitle": "Everything you need to know about applying to inVision U",
    "faq.majors_title": "Programs",
    "faq.req_title": "Minimum Requirements",
    "faq.back": "Back",
    "faq.apply": "Apply Now",
  },
  ru: {
    "faq.title": "Часто задаваемые вопросы",
    "faq.subtitle": "Всё, что нужно знать о поступлении в inVision U",
    "faq.majors_title": "Программы",
    "faq.req_title": "Минимальные требования",
    "faq.back": "Назад",
    "faq.apply": "Подать заявку",
  },
  kk: {
    "faq.title": "Жиі қойылатын сұрақтар",
    "faq.subtitle": "inVision U-ға өтінім беру туралы білуіңіз керек барлық нәрсе",
    "faq.majors_title": "Бағдарламалар",
    "faq.req_title": "Минималды талаптар",
    "faq.back": "Артқа",
    "faq.apply": "Өтінім беру",
  },
};

export default function FAQPage() {
  const [lang, setLang] = useState<Lang>("ru");
  const [openIdx, setOpenIdx] = useState<number | null>(null);
  const { theme, toggleTheme } = useTheme();

  const t = (key: string) => UI_LABELS[lang]?.[key] ?? key;
  const faqs = FAQ_DATA[lang];
  const reqs = REQUIREMENTS[lang];
  const majors = MAJORS_INFO[lang];

  return (
    <div className="min-h-screen bg-background">
      {/* Top bar */}
      <div className="border-b border-border px-6 py-3 flex items-center justify-between">
        <Link href="/" className="flex items-center gap-2 text-muted-foreground hover:text-foreground text-sm transition-colors">
          <ArrowLeft size={16} /> {t("faq.back")}
        </Link>
        <div className="flex gap-2" role="group" aria-label="Language selection">
          {(["en", "ru", "kk"] as const).map((l) => (
            <button
              key={l}
              onClick={() => setLang(l)}
              aria-pressed={lang === l}
              className={`text-xs px-3 py-1.5 rounded-full border font-medium transition-colors ${
                lang === l
                  ? "border-primary font-bold"
                  : "border-border text-muted-foreground hover:border-primary"
              }`}
              style={lang === l ? { backgroundColor: "#c1f11d", borderColor: "#c1f11d", color: "#111" } : undefined}
            >
              {l.toUpperCase()}
            </button>
          ))}
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={toggleTheme}
            className="p-2 rounded-lg border border-border text-muted-foreground hover:text-foreground transition-colors"
            aria-label="Toggle theme"
          >
            {theme === "light" ? <Moon size={16} /> : <Sun size={16} />}
          </button>
          <Link href="/apply">
            <button
              className="text-sm px-4 py-2 rounded-lg font-semibold transition-colors"
              style={{ backgroundColor: "#c1f11d", color: "#111" }}
            >
              {t("faq.apply")}
            </button>
          </Link>
        </div>
      </div>

      <div className="max-w-3xl mx-auto px-4 py-10 space-y-12">
        {/* Header */}
        <div className="text-center space-y-2">
          <div className="flex justify-center mb-4">
            <div className="w-12 h-12 rounded-2xl flex items-center justify-center" style={{ backgroundColor: "#c1f11d" }}>
              <HelpCircle size={24} className="text-black" />
            </div>
          </div>
          <h1 className="text-3xl font-bold text-foreground">{t("faq.title")}</h1>
          <p className="text-muted-foreground">{t("faq.subtitle")}</p>
        </div>

        {/* Programs */}
        <section aria-labelledby="programs-heading">
          <div className="flex items-center gap-2 mb-4">
            <BookOpen size={18} className="text-primary" />
            <h2 id="programs-heading" className="text-lg font-semibold text-foreground">{t("faq.majors_title")}</h2>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {majors.map((m) => (
              <div
                key={m.tag}
                className="flex items-start gap-3 p-4 rounded-xl border border-border bg-card hover:border-primary/50 transition-colors"
              >
                <span
                  className="text-[10px] font-bold px-2 py-1 rounded-lg shrink-0 mt-0.5"
                  style={{ backgroundColor: "#c1f11d", color: "#111" }}
                >
                  {m.tag}
                </span>
                <span className="text-sm text-foreground leading-snug">{m.name}</span>
              </div>
            ))}
          </div>
        </section>

        {/* Requirements */}
        <section aria-labelledby="requirements-heading">
          <h2 id="requirements-heading" className="text-lg font-semibold text-foreground mb-4">{reqs.heading}</h2>
          <div className="space-y-3">
            {reqs.items.map((item) => (
              <div key={item.number} className="flex gap-4 p-4 rounded-xl border border-border bg-card">
                <div
                  className="text-2xl font-black shrink-0 leading-none mt-0.5"
                  style={{ color: "#c1f11d" }}
                  aria-hidden="true"
                >
                  {item.number}
                </div>
                <div>
                  <h3 className="font-semibold text-foreground mb-1">{item.title}</h3>
                  <p className="text-sm text-muted-foreground whitespace-pre-line">{item.content}</p>
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* FAQ Accordion */}
        <section aria-labelledby="faq-heading">
          <h2 id="faq-heading" className="text-lg font-semibold text-foreground mb-4">{t("faq.title")}</h2>
          <div className="space-y-2" role="list">
            {faqs.map((item, i) => (
              <div
                key={i}
                className="border border-border rounded-xl bg-card overflow-hidden"
                role="listitem"
              >
                <button
                  id={`faq-btn-${i}`}
                  aria-expanded={openIdx === i}
                  aria-controls={`faq-panel-${i}`}
                  onClick={() => setOpenIdx(openIdx === i ? null : i)}
                  className="w-full text-left flex items-center justify-between px-5 py-4 gap-3 hover:bg-muted/40 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                >
                  <span className="text-sm font-medium text-foreground">{item.q}</span>
                  <ChevronDown
                    size={16}
                    className={`shrink-0 text-muted-foreground transition-transform duration-200 ${openIdx === i ? "rotate-180" : ""}`}
                    aria-hidden="true"
                  />
                </button>
                <div
                  id={`faq-panel-${i}`}
                  role="region"
                  aria-labelledby={`faq-btn-${i}`}
                  className={`transition-all duration-200 ease-in-out overflow-hidden ${openIdx === i ? "max-h-[600px]" : "max-h-0"}`}
                >
                  <div className="px-5 pb-4 text-sm text-muted-foreground">
                    {item.html ? (
                      <span dangerouslySetInnerHTML={{ __html: item.a }} />
                    ) : (
                      item.a
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* CTA */}
        <div className="text-center py-6 border-t border-border">
          <p className="text-muted-foreground mb-4 text-sm">
            {lang === "ru" ? "Готовы? Подайте заявку уже сегодня." :
             lang === "kk" ? "Дайынсыз ба? Бүгін өтінім беріңіз." :
             "Ready to apply?"}
          </p>
          <Link href="/apply">
            <button
              className="px-8 py-3 rounded-xl font-semibold text-sm transition-colors hover:opacity-90"
              style={{ backgroundColor: "#c1f11d", color: "#111" }}
            >
              {t("faq.apply")} →
            </button>
          </Link>
        </div>
      </div>
    </div>
  );
}
