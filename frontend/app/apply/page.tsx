"use client";

import { useState, FormEvent, useRef, useEffect } from "react";
import Link from "next/link";
import axios from "axios";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { CheckCircle, ArrowLeft, Upload, AlertTriangle, X, ChevronDown, User, Phone, GraduationCap, ClipboardList } from "lucide-react";

const publicApi = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api",
});

interface MajorOption {
  tag: string;
  en: string;
  ru: string;
  kk: string;
}

const SECTIONS = [
  { id: "personal", label: "Personal Information", icon: User },
  { id: "contact", label: "Contact Information", icon: Phone },
  { id: "education", label: "Education", icon: GraduationCap },
  { id: "test", label: "Internal Test", icon: ClipboardList },
] as const;
type SectionId = (typeof SECTIONS)[number]["id"];

// Countries with flag emojis
const COUNTRIES = [
  { code: "KZ", name: "Kazakhstan", flag: "\u{1F1F0}\u{1F1FF}" },
  { code: "RU", name: "Russia", flag: "\u{1F1F7}\u{1F1FA}" },
  { code: "UZ", name: "Uzbekistan", flag: "\u{1F1FA}\u{1F1FF}" },
  { code: "KG", name: "Kyrgyzstan", flag: "\u{1F1F0}\u{1F1EC}" },
  { code: "TJ", name: "Tajikistan", flag: "\u{1F1F9}\u{1F1EF}" },
  { code: "TM", name: "Turkmenistan", flag: "\u{1F1F9}\u{1F1F2}" },
  { code: "CN", name: "China", flag: "\u{1F1E8}\u{1F1F3}" },
  { code: "IN", name: "India", flag: "\u{1F1EE}\u{1F1F3}" },
  { code: "US", name: "United States", flag: "\u{1F1FA}\u{1F1F8}" },
  { code: "GB", name: "United Kingdom", flag: "\u{1F1EC}\u{1F1E7}" },
  { code: "DE", name: "Germany", flag: "\u{1F1E9}\u{1F1EA}" },
  { code: "FR", name: "France", flag: "\u{1F1EB}\u{1F1F7}" },
  { code: "TR", name: "Turkey", flag: "\u{1F1F9}\u{1F1F7}" },
  { code: "AE", name: "United Arab Emirates", flag: "\u{1F1E6}\u{1F1EA}" },
  { code: "KR", name: "South Korea", flag: "\u{1F1F0}\u{1F1F7}" },
  { code: "JP", name: "Japan", flag: "\u{1F1EF}\u{1F1F5}" },
  { code: "AZ", name: "Azerbaijan", flag: "\u{1F1E6}\u{1F1FF}" },
  { code: "GE", name: "Georgia", flag: "\u{1F1EC}\u{1F1EA}" },
  { code: "AM", name: "Armenia", flag: "\u{1F1E6}\u{1F1F2}" },
  { code: "BY", name: "Belarus", flag: "\u{1F1E7}\u{1F1FE}" },
  { code: "UA", name: "Ukraine", flag: "\u{1F1FA}\u{1F1E6}" },
  { code: "PL", name: "Poland", flag: "\u{1F1F5}\u{1F1F1}" },
  { code: "MN", name: "Mongolia", flag: "\u{1F1F2}\u{1F1F3}" },
  { code: "AF", name: "Afghanistan", flag: "\u{1F1E6}\u{1F1EB}" },
  { code: "PK", name: "Pakistan", flag: "\u{1F1F5}\u{1F1F0}" },
  { code: "BD", name: "Bangladesh", flag: "\u{1F1E7}\u{1F1E9}" },
  { code: "ID", name: "Indonesia", flag: "\u{1F1EE}\u{1F1E9}" },
  { code: "MY", name: "Malaysia", flag: "\u{1F1F2}\u{1F1FE}" },
  { code: "BR", name: "Brazil", flag: "\u{1F1E7}\u{1F1F7}" },
  { code: "MX", name: "Mexico", flag: "\u{1F1F2}\u{1F1FD}" },
  { code: "CA", name: "Canada", flag: "\u{1F1E8}\u{1F1E6}" },
  { code: "AU", name: "Australia", flag: "\u{1F1E6}\u{1F1FA}" },
  { code: "IT", name: "Italy", flag: "\u{1F1EE}\u{1F1F9}" },
  { code: "ES", name: "Spain", flag: "\u{1F1EA}\u{1F1F8}" },
  { code: "SA", name: "Saudi Arabia", flag: "\u{1F1F8}\u{1F1E6}" },
  { code: "EG", name: "Egypt", flag: "\u{1F1EA}\u{1F1EC}" },
  { code: "NG", name: "Nigeria", flag: "\u{1F1F3}\u{1F1EC}" },
  { code: "ZA", name: "South Africa", flag: "\u{1F1FF}\u{1F1E6}" },
  { code: "OTHER", name: "Other", flag: "\u{1F30D}" },
];

const PERSONALITY_QUESTIONS = [
  {
    q: "1. When I see that something can be improved, I usually\u2026",
    options: [
      "Wait until I fully understand how to make it better.",
      "Discuss with others whether something should be changed.",
      "Try to make an improvement right away, even a small one.",
      "Notice what could be better but avoid getting involved unless asked.",
    ],
  },
  {
    q: "2. When a group discussion reaches a \u201cdead end\u201d and no one suggests ideas, I\u2026",
    options: [
      "Pause and wait for someone else to speak first.",
      "Ask a question to restart the discussion.",
      "Offer an idea, even if it\u2019s not fully developed.",
      "Suggest postponing the discussion until later.",
    ],
  },
  {
    q: "3. When I face a problem I\u2019ve never encountered before, I\u2026",
    options: [
      "Look for a solution on my own, using available resources.",
      "First learn how others have solved similar problems.",
      "Ask someone who knows better.",
      "Leave it as it is to avoid making things worse.",
    ],
  },
  {
    q: "4. If a teammate makes a mistake that affects my part of the project, I\u2026",
    options: [
      "Discuss it with them and help fix it since it\u2019s a shared outcome.",
      "Inform the supervisor and let them decide what to do.",
      "Mention the mistake but avoid interfering \u2014 everyone is responsible for their part.",
      "Point out that my part was done correctly and it\u2019s not my responsibility.",
    ],
  },
  {
    q: "5. If I promised to help someone but realize I won\u2019t make it on time, I\u2026",
    options: [
      "Let them know I can\u2019t help and explain why.",
      "Inform them early, apologize, and try to fulfill my promise later.",
      "Avoid contact for a while to skip awkward explanations.",
      "Wait until they solve it on their own.",
    ],
  },
  {
    q: "6. If I promised the team to complete a task but realize I can\u2019t, I\u2026",
    options: [
      "Inform them in advance and help reassign the task.",
      "Stay silent, hoping to finish at least part of it.",
      "Explain later that my situation changed and it\u2019s not my fault.",
      "Warn the team at the last moment and ask for their help.",
    ],
  },
  {
    q: "7. When I\u2019m learning something new and face difficulties, I\u2026",
    options: [
      "See it as part of growth and look for solutions.",
      "Stop doing it, thinking it\u2019s just not for me.",
      "Put it aside and return later when inspired.",
      "Ask for help to save time.",
    ],
  },
  {
    q: "8. When I think about my future, I\u2026",
    options: [
      "Plan which skills to develop to move forward.",
      "Go with the flow \u2014 life will show the way.",
      "Listen to advice and try to define a direction for growth.",
      "Don\u2019t think about it much \u2014 luck decides most things.",
    ],
  },
  {
    q: "9. When I notice the teacher explains something superficially, I\u2026",
    options: [
      "Assume it\u2019s not that important.",
      "Ask for additional reading materials.",
      "Wait, hoping the next class will clarify things.",
      "Find extra sources and study the topic deeper on my own.",
    ],
  },
  {
    q: "10. When a task takes more time than expected, I\u2026",
    options: [
      "Adjust my approach and continue until I finish.",
      "Switch to something else to avoid getting stuck.",
      "Wait until I get more support or resources.",
      "Break it down into smaller parts to move gradually.",
    ],
  },
  {
    q: "11. When I lack motivation or energy, I\u2026",
    options: [
      "Remind myself why I started and finish anyway.",
      "Switch to easier parts to maintain momentum.",
      "Stop, believing that forcing myself won\u2019t help.",
      "Take a break and wait for the right mood.",
    ],
  },
  {
    q: "12. When competition rules suddenly change at the final stage, I\u2026",
    options: [
      "Complete it formally \u2014 it\u2019s too late to change much.",
      "Keep the same approach but improve the presentation.",
      "Adapt my project to the new requirements.",
      "Drop out, believing it\u2019s unfair.",
    ],
  },
  {
    q: "13. When I\u2019m praised for good work, I\u2026",
    options: [
      "Take it as a sign to aim even higher.",
      "Feel happy and think about maintaining this level.",
      "Thank them and keep doing the same.",
      "Feel embarrassed and try not to stand out again.",
    ],
  },
  {
    q: "14. When I\u2019m offered a challenging, high-risk task where I could shine, I\u2026",
    options: [
      "Avoid such situations.",
      "Weigh the risks and prepare to minimize errors.",
      "Wait for someone else to take it first.",
      "Take it on \u2014 the chance to prove myself is worth the risk.",
    ],
  },
  {
    q: "15. When I imagine my ideal career path, I see myself\u2026",
    options: [
      "In a position where my decisions make a real impact.",
      "As an expert whose opinion is respected.",
      "In a stable job with little responsibility.",
      "In a calm job without pressure to prove myself.",
    ],
  },
  {
    q: "16. When I see injustice, but speaking up might complicate my life, I\u2026",
    options: [
      "Prefer not to get involved to avoid trouble.",
      "Still speak up, because it\u2019s hard for me to stay silent when someone is hurt.",
      "Think it\u2019s not my business since I can\u2019t change anything anyway.",
      "Try at least to talk to the person affected.",
    ],
  },
  {
    q: "17. If I notice someone takes my help for granted, I\u2026",
    options: [
      "Try to explain that my help is a choice, not an obligation.",
      "Feel irritated and cut off contact completely.",
      "Distance myself, as I don\u2019t want to be taken advantage of.",
      "Keep helping anyway \u2014 I don\u2019t do it for gratitude.",
    ],
  },
  {
    q: "18. When I see someone doesn\u2019t have the materials for an exam, I\u2026",
    options: [
      "Pretend not to notice, otherwise they\u2019ll keep asking.",
      "Share mine \u2014 we\u2019re learning together after all.",
      "Tell them where they can find the materials.",
      "Think everyone should take care of themselves.",
    ],
  },
  {
    q: "19. If, during an internship, I see that the supervisor uses students as free labor, I\u2026",
    options: [
      "Gather opinions from other interns and prepare a constructive message to management together.",
      "Agree with others that \u201ceveryone does that, there is no point trying to change anything.\u201d",
      "Prefer to finish my internship quietly and avoid unnecessary conflicts.",
      "Try to have a calm, honest conversation with the supervisor personally.",
    ],
  },
  {
    q: "20. When I notice that my university deletes negative feedback, explaining it as \u201cprotecting its image,\u201d I\u2026",
    options: [
      "Think it\u2019s normal practice and see nothing wrong with it.",
      "Ask the administration if they could at least analyze the complaints before deleting them.",
      "Suggest responding openly to criticism and using it to fix problems.",
      "Decide not to get involved, thinking it\u2019s not my area of influence.",
    ],
  },
  {
    q: "21. When I see recycling bins installed on campus but nobody uses them, I\u2026",
    options: [
      "Suggest adding a short explanation nearby to raise awareness.",
      "Think people just need time to adapt and decide not to interfere.",
      "Ignore it \u2014 it\u2019s not my initiative or responsibility.",
      "Simply use the bin correctly myself, because personal example matters more than words.",
    ],
  },
  {
    q: "22. If I am assigned to work with someone who is considered \u201cdifficult\u201d by others, I\u2026",
    options: [
      "Stay calm and cooperate without unnecessary emotions.",
      "Try to see their strengths and understand what lies behind their behavior.",
      "Work with them formally, just to complete the task.",
      "Prefer to avoid collaboration to save my nerves.",
    ],
  },
  {
    q: "23. When someone shares a personal story that goes against my values, I\u2026",
    options: [
      "Decide that I shouldn\u2019t listen to something that contradicts my beliefs.",
      "Politely change the topic if I feel uncomfortable.",
      "Try to avoid such topics in the future.",
      "Try to listen without judgment to better understand the person.",
    ],
  },
  {
    q: "24. When someone receives fewer opportunities than others, I\u2026",
    options: [
      "Try to understand how to level the playing field or support them.",
      "Believe that if someone didn\u2019t get the opportunity, they probably didn\u2019t deserve it.",
      "Think the main thing is to focus on my own business.",
      "Think life isn\u2019t always fair, but I try to be attentive to people.",
    ],
  },
  {
    q: "25. If I find out that someone is getting advantages through connections, I\u2026",
    options: [
      "Consider that being able to \u201cnegotiate\u201d is also a kind of skill.",
      "Discuss it with others who are also dissatisfied, to see if something can be done.",
      "Think such things have always existed and are part of the system.",
      "Believe it\u2019s unfair and at least try to speak openly about it.",
    ],
  },
  {
    q: "26. When I realize that my inaction has caused someone a problem, I\u2026",
    options: [
      "Think that if I didn\u2019t do anything, I\u2019m not to blame.",
      "Admit that I\u2019m also responsible, because harm can come from silence.",
      "Believe that everyone is only responsible for their own actions.",
      "Reflect on how I could have acted differently.",
    ],
  },
  {
    q: "27. When someone says \u201cthe end justifies the means,\u201d I\u2026",
    options: [
      "Fully agree, because winners aren\u2019t judged.",
      "Partially agree \u2014 it depends on the consequences.",
      "Think that sometimes tough decisions are unavoidable.",
      "Believe that if the means cause harm, the goal loses its value.",
    ],
  },
  {
    q: "28. When I think about my city (or neighborhood), I\u2026",
    options: [
      "Dream of leaving as soon as possible \u2014 nothing will ever change here.",
      "Wish it could become more comfortable and safer.",
      "Think that little depends on me personally.",
      "Want people here to feel they can make changes themselves.",
    ],
  },
  {
    q: "29. When I think about my future, I\u2026",
    options: [
      "Believe that stability and comfort are the most important things.",
      "Want to combine personal success with helping others.",
      "Want to be useful to people, not just earn money.",
      "Think that helping others is a personal choice, not an obligation.",
    ],
  },
  {
    q: "30. When I hear someone criticize our community, I\u2026",
    options: [
      "Think about how to reduce the reasons for criticism.",
      "Just agree \u2014 everyone has their opinion, no need to take it personally.",
      "Try to explain that things are not so simple and highlight the positive sides.",
      "Ignore it \u2014 criticism always exists, there\u2019s no point in trying to change anything.",
    ],
  },
  {
    q: "31. When I notice that someone in the team has withdrawn, I\u2026",
    options: [
      "Try to gently engage them through open discussion.",
      "Approach them directly and ask what\u2019s going on, as it\u2019s important to bring them back into the process.",
      "Think that not everyone has to be equally active.",
      "Don\u2019t pay attention \u2014 if they\u2019re silent, it means they don\u2019t want to participate.",
    ],
  },
  {
    q: "32. When a team discussion reaches a dead end, I\u2026",
    options: [
      "Ask for an outside perspective to get a fresh view.",
      "Suggest identifying exactly where we disagree and moving forward from there.",
      "Propose voting to settle the issue and move on.",
      "Believe there\u2019s no point in arguing any further.",
    ],
  },
  {
    q: "33. When my classmates suggest teaming up for a project or competition, I\u2026",
    options: [
      "Think it\u2019s easier to do everything myself so I don\u2019t depend on others.",
      "Agree to join if I understand how the roles will be distributed.",
      "Feel glad to work together, since collaboration increases our chances of success.",
      "Support the idea but try not to take an active role.",
    ],
  },
  {
    q: "34. When someone in the group doesn\u2019t keep their promises, I\u2026",
    options: [
      "Calmly discuss what went wrong and look for ways to rebuild trust.",
      "Decide that I can\u2019t trust anyone and it\u2019s better to do everything myself.",
      "Feel disappointed but try not to show it.",
      "Bring it up directly but politely, without blaming.",
    ],
  },
  {
    q: "35. When I come across new information, I\u2026",
    options: [
      "Trust it if the source looks reliable.",
      "Sometimes verify it if something feels off.",
      "Think about who said it and why, and try to understand their motives and source.",
      "Don\u2019t see the point in checking \u2014 you can never know the full truth anyway.",
    ],
  },
  {
    q: "36. When someone points out my mistake, I\u2026",
    options: [
      "Listen carefully, since others can notice what I might have missed.",
      "Assume the person is just being picky.",
      "Try not to pay attention, though I don\u2019t like being criticized.",
      "Feel a bit irritated but try to take the feedback constructively.",
    ],
  },
  {
    q: "37. When I\u2019m asked to do something new, I\u2026",
    options: [
      "Refuse, because I don\u2019t like feeling inexperienced.",
      "Accept that I might not succeed right away and stay calm about it.",
      "Prefer to stick to what I already know well.",
      "Hope I\u2019ll manage, though I worry about possible mistakes.",
    ],
  },
  {
    q: "38. When I think about my future, I\u2026",
    options: [
      "Believe that my efforts will determine what I achieve.",
      "Think that circumstances and luck play the biggest role.",
      "Hope that if I don\u2019t give up, things will work out.",
      "Believe success depends more on fate than on personal effort.",
    ],
  },
  {
    q: "39. When I feel envy, I\u2026",
    options: [
      "Acknowledge it and think about what I can do to improve myself.",
      "Think life is unfair and get into a bad mood.",
      "Try not to show that something has affected me.",
      "Shift my focus to something else to calm down.",
    ],
  },
  {
    q: "40. When I feel irritated by others, I\u2026",
    options: [
      "Step aside to cool down a bit.",
      "Try not to show my irritation, even if I\u2019m boiling inside.",
      "Acknowledge my feelings and try to understand what exactly triggered me.",
      "Think irritation is a normal reaction and don\u2019t see a reason to overthink it.",
    ],
  },
];

export default function ApplyPage() {
  const [majors, setMajors] = useState<MajorOption[]>([]);
  const [activeSection, setActiveSection] = useState<SectionId>("personal");
  const [form, setForm] = useState({
    first_name: "", last_name: "", patronymic: "", email: "", phone: "", telegram: "",
    date_of_birth: "", gender: "", nationality: "Kazakhstan", iin: "",
    home_country: "Kazakhstan", city: "", instagram: "", whatsapp: "",
    school: "", major: "", youtube_url: "",
    exam_type: "", ielts_score: "", toefl_score: "",
    certificate_type: "",
    achievements: "", extracurriculars: "", essay: "", motivation_statement: "",
    disability: "",
  });
  const [personalityAnswers, setPersonalityAnswers] = useState<Record<number, number>>({});
  const [agreePrivacy, setAgreePrivacy] = useState(false);
  const [agreeAge, setAgreeAge] = useState(false);
  const [photo, setPhoto] = useState<File | null>(null);
  const [photoPreview, setPhotoPreview] = useState<string | null>(null);
  const [englishCertFile, setEnglishCertFile] = useState<File | null>(null);
  const [certificateFile, setCertificateFile] = useState<File | null>(null);
  const [additionalDocsFile, setAdditionalDocsFile] = useState<File | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState("");
  const fileRef = useRef<HTMLInputElement>(null);
  const englishCertRef = useRef<HTMLInputElement>(null);
  const certificateRef = useRef<HTMLInputElement>(null);
  const additionalDocsRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    publicApi.get<MajorOption[]>("/majors").then((r) => setMajors(r.data)).catch(() => {});
  }, []);

  const update = (field: string, value: string) => setForm({ ...form, [field]: value });

  const handlePhotoChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    if (file.size > 5 * 1024 * 1024) { setError("Photo must be under 5 MB"); return; }
    setPhoto(file);
    setPhotoPreview(URL.createObjectURL(file));
    setError("");
  };

  const handleFileChange = (
    e: React.ChangeEvent<HTMLInputElement>,
    setter: (f: File | null) => void
  ) => {
    const file = e.target.files?.[0];
    if (!file) return;
    if (file.size > 10 * 1024 * 1024) { setError("File must be under 10 MB"); return; }
    setter(file);
    setError("");
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");

    if (!form.last_name || !form.first_name) { setError("First name and last name are required"); return; }
    if (!form.email) { setError("Email is required"); return; }
    if (!form.date_of_birth) { setError("Date of birth is required"); return; }
    if (!form.gender) { setError("Gender is required"); return; }
    if (!form.nationality) { setError("Nationality is required"); return; }
    if (!form.iin || form.iin.length !== 12) { setError("IIN must be exactly 12 digits"); return; }
    if (!form.phone) { setError("Phone number is required"); return; }
    if (!form.telegram) { setError("Telegram is required"); return; }
    if (!form.home_country) { setError("Country is required"); return; }
    if (!form.city) { setError("City is required"); return; }
    if (!form.youtube_url) { setError("YouTube presentation link is required"); return; }
    if (!form.exam_type) { setError("Please select IELTS or TOEFL"); return; }
    if (form.exam_type === "IELTS" && (!form.ielts_score || parseFloat(form.ielts_score) < 6.0)) {
      setError("IELTS score must be at least 6.0"); return;
    }
    if (form.exam_type === "TOEFL" && (!form.toefl_score || parseInt(form.toefl_score) < 60)) {
      setError("TOEFL score must be at least 60"); return;
    }
    if (!englishCertFile) { setError("English proficiency certificate is required"); return; }
    if (!form.certificate_type) { setError("Please select certificate type (UNT or NIS)"); return; }
    if (!certificateFile) { setError("Certificate file is required"); return; }
    if (!form.essay || form.essay.length < 50) { setError("Essay must be at least 50 characters"); return; }
    if (Object.keys(personalityAnswers).length < 40) { setError("Please answer all 40 personality test questions"); return; }
    if (!agreePrivacy || !agreeAge) { setError("You must agree to both consent statements"); return; }

    setSubmitting(true);
    try {
      const fullName = `${form.last_name} ${form.first_name}`.trim();
      const answersJson = JSON.stringify(personalityAnswers);

      const res = await publicApi.post<{ id: number }>("/apply", {
        full_name: fullName,
        first_name: form.first_name,
        last_name: form.last_name,
        patronymic: form.patronymic || null,
        email: form.email,
        phone: form.phone,
        telegram: form.telegram,
        date_of_birth: form.date_of_birth || null,
        gender: form.gender || null,
        nationality: form.nationality || null,
        iin: form.iin || null,
        home_country: form.home_country || null,
        city: form.city || null,
        instagram: form.instagram || null,
        whatsapp: form.whatsapp || null,
        school: form.school || null,
        major: form.major || null,
        youtube_url: form.youtube_url,
        exam_type: form.exam_type || null,
        ielts_score: form.ielts_score ? parseFloat(form.ielts_score) : null,
        toefl_score: form.toefl_score ? parseInt(form.toefl_score) : null,
        certificate_type: form.certificate_type || null,
        achievements: form.achievements || null,
        extracurriculars: form.extracurriculars || null,
        essay: form.essay,
        motivation_statement: form.motivation_statement || null,
        disability: form.disability || null,
        personality_answers: answersJson,
      });

      const candId = res.data.id;

      // Upload files in parallel
      const uploads: Promise<void>[] = [];
      if (photo) {
        const fd = new FormData();
        fd.append("photo", photo);
        uploads.push(publicApi.post(`/candidates/${candId}/photo`, fd, { headers: { "Content-Type": "multipart/form-data" } }).then(() => {}));
      }
      if (englishCertFile) {
        const fd = new FormData();
        fd.append("file", englishCertFile);
        uploads.push(publicApi.post(`/candidates/${candId}/document/english_cert`, fd, { headers: { "Content-Type": "multipart/form-data" } }).then(() => {}));
      }
      if (certificateFile) {
        const fd = new FormData();
        fd.append("file", certificateFile);
        uploads.push(publicApi.post(`/candidates/${candId}/document/certificate`, fd, { headers: { "Content-Type": "multipart/form-data" } }).then(() => {}));
      }
      if (additionalDocsFile) {
        const fd = new FormData();
        fd.append("file", additionalDocsFile);
        uploads.push(publicApi.post(`/candidates/${candId}/document/additional_docs`, fd, { headers: { "Content-Type": "multipart/form-data" } }).then(() => {}));
      }
      await Promise.allSettled(uploads);

      setSuccess(true);
    } catch (err) {
      if (axios.isAxiosError(err) && err.response?.status === 409) {
        setError("This email is already registered.");
      } else {
        setError("Something went wrong. Please try again.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (success) {
    return (
      <div className="min-h-screen bg-white flex items-center justify-center p-4">
        <Card className="max-w-md w-full border-green-200 bg-green-50">
          <CardContent className="p-10 text-center">
            <CheckCircle className="mx-auto text-green-500 mb-4" size={52} />
            <h2 className="text-2xl font-bold text-green-800">Application Submitted!</h2>
            <p className="text-green-700 mt-2 text-sm">
              Thank you for applying to inVision U. We will review your application and contact you via email.
            </p>
            <Link href="/">
              <Button variant="outline" className="mt-6">
                <ArrowLeft size={16} className="mr-2" /> Back to Home
              </Button>
            </Link>
          </CardContent>
        </Card>
      </div>
    );
  }

  const answeredCount = Object.keys(personalityAnswers).length;

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b sticky top-0 z-10">
        <div className="max-w-3xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Link href="/" className="text-gray-400 hover:text-gray-600">
              <ArrowLeft size={20} />
            </Link>
            <div>
              <h1 className="text-xl font-bold text-gray-900">inVision U</h1>
              <p className="text-xs text-gray-500">Application Form</p>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-3xl mx-auto px-4 py-6">
        {/* Section tabs */}
        <div className="flex gap-1 mb-6 bg-white rounded-xl p-1 border overflow-x-auto">
          {SECTIONS.map((s) => {
            const Icon = s.icon;
            const isActive = activeSection === s.id;
            return (
              <button
                key={s.id}
                onClick={() => setActiveSection(s.id)}
                className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-medium transition-all whitespace-nowrap flex-1 justify-center ${
                  isActive
                    ? "text-gray-900 shadow-sm"
                    : "text-gray-500 hover:text-gray-700"
                }`}
                style={isActive ? { backgroundColor: "#c1f11d" } : undefined}
              >
                <Icon size={16} />
                <span className="hidden sm:inline">{s.label}</span>
              </button>
            );
          })}
        </div>

        <form onSubmit={handleSubmit} noValidate>
          <Card className="border bg-white mb-6">
            <CardContent className="p-6">
              {/* ═══ SECTION 1: Personal Information ═══ */}
              {activeSection === "personal" && (
                <div className="space-y-6">
                  <h2 className="text-lg font-bold text-gray-900">Personal Information</h2>

                  <div>
                    <h3 className="text-sm font-semibold text-gray-700 mb-3">Applicant details</h3>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <Field label="Last Name *" htmlFor="last_name">
                        <Input id="last_name" value={form.last_name} onChange={(e) => update("last_name", e.target.value)} required />
                      </Field>
                      <Field label="First Name *" htmlFor="first_name">
                        <Input id="first_name" value={form.first_name} onChange={(e) => update("first_name", e.target.value)} required />
                      </Field>
                      <Field label="Patronymic" htmlFor="patronymic">
                        <Input id="patronymic" value={form.patronymic} onChange={(e) => update("patronymic", e.target.value)} />
                      </Field>
                      <Field label="Date of Birth *" htmlFor="date_of_birth">
                        <Input id="date_of_birth" type="date" value={form.date_of_birth} onChange={(e) => update("date_of_birth", e.target.value)} required />
                      </Field>
                      <Field label="Gender *" htmlFor="gender">
                        <div className="relative">
                          <select value={form.gender} onChange={(e) => update("gender", e.target.value)} required
                            className="w-full appearance-none bg-white border border-input rounded-lg px-3 py-2.5 pr-10 text-sm focus:outline-none focus:ring-2 focus:ring-ring">
                            <option value="">Select gender...</option>
                            <option value="Male">Male</option>
                            <option value="Female">Female</option>
                          </select>
                          <ChevronDown size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none" />
                        </div>
                      </Field>
                      <Field label="Email *" htmlFor="email">
                        <Input id="email" type="email" value={form.email} onChange={(e) => update("email", e.target.value)} required />
                      </Field>
                    </div>
                  </div>

                  <div>
                    <h3 className="text-sm font-semibold text-gray-700 mb-3">Nationality and passport details</h3>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <Field label="Citizenship *" htmlFor="nationality">
                        <div className="relative">
                          <select value={form.nationality} onChange={(e) => update("nationality", e.target.value)} required
                            className="w-full appearance-none bg-white border border-input rounded-lg px-3 py-2.5 pr-10 text-sm focus:outline-none focus:ring-2 focus:ring-ring">
                            {COUNTRIES.map((c) => (
                              <option key={c.code} value={c.name}>{c.flag} {c.name}</option>
                            ))}
                          </select>
                          <ChevronDown size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none" />
                        </div>
                      </Field>
                      <Field label="Individual Identification Number (IIN) *" htmlFor="iin">
                        <Input id="iin" value={form.iin} onChange={(e) => update("iin", e.target.value.replace(/\D/g, "").slice(0, 12))}
                          placeholder="123456789012" maxLength={12} required />
                        <p className="text-xs text-gray-400 mt-1">{form.iin.length}/12 digits</p>
                      </Field>
                    </div>
                  </div>

                  {/* Photo upload */}
                  <div>
                    <h3 className="text-sm font-semibold text-gray-700 mb-2">Profile Photo</h3>
                    <div className="flex items-start gap-4">
                      {photoPreview ? (
                        <div className="relative">
                          <img src={photoPreview} alt="Preview" className="w-24 h-24 rounded-xl object-cover border" />
                          <button type="button" onClick={() => { setPhoto(null); setPhotoPreview(null); }}
                            className="absolute -top-2 -right-2 bg-red-500 text-white rounded-full p-0.5">
                            <X size={12} />
                          </button>
                        </div>
                      ) : (
                        <button type="button" onClick={() => fileRef.current?.click()}
                          className="w-24 h-24 rounded-xl border-2 border-dashed flex flex-col items-center justify-center gap-1 text-gray-400 hover:border-gray-400 hover:text-gray-600 transition-colors">
                          <Upload size={20} />
                          <span className="text-[10px]">Upload</span>
                        </button>
                      )}
                      <p className="text-xs text-gray-500">Max 5 MB, JPEG or PNG</p>
                    </div>
                    <input ref={fileRef} type="file" accept="image/*" className="hidden" onChange={handlePhotoChange} />
                  </div>

                  {/* Major */}
                  <Field label="Preferred Major *" htmlFor="major">
                    <div className="relative">
                      <select value={form.major} onChange={(e) => update("major", e.target.value)} required
                        className="w-full appearance-none bg-white border border-input rounded-lg px-3 py-2.5 pr-10 text-sm focus:outline-none focus:ring-2 focus:ring-ring">
                        <option value="">Select a major...</option>
                        {majors.map((m) => (
                          <option key={m.tag} value={m.tag}>{m.tag} — {m.en}</option>
                        ))}
                      </select>
                      <ChevronDown size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none" />
                    </div>
                  </Field>

                  {/* Disability */}
                  <Field label="Disability / Accessibility Needs" htmlFor="disability">
                    <p className="text-xs text-gray-500 mb-2">Optional. Not used in scoring.</p>
                    <Textarea id="disability" rows={2} value={form.disability} onChange={(e) => update("disability", e.target.value)} />
                  </Field>
                </div>
              )}

              {/* ═══ SECTION 2: Contact Information ═══ */}
              {activeSection === "contact" && (
                <div className="space-y-6">
                  <h2 className="text-lg font-bold text-gray-900">Contact Information</h2>

                  <div>
                    <h3 className="text-sm font-semibold text-gray-700 mb-3">Home Address</h3>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <Field label="Country *" htmlFor="home_country">
                        <div className="relative">
                          <select value={form.home_country} onChange={(e) => update("home_country", e.target.value)} required
                            className="w-full appearance-none bg-white border border-input rounded-lg px-3 py-2.5 pr-10 text-sm focus:outline-none focus:ring-2 focus:ring-ring">
                            {COUNTRIES.map((c) => (
                              <option key={c.code} value={c.name}>{c.flag} {c.name}</option>
                            ))}
                          </select>
                          <ChevronDown size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none" />
                        </div>
                      </Field>
                      <Field label="City *" htmlFor="city">
                        <Input id="city" value={form.city} onChange={(e) => update("city", e.target.value)} required />
                      </Field>
                    </div>
                  </div>

                  <div>
                    <h3 className="text-sm font-semibold text-gray-700 mb-3">Contact details</h3>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <Field label="Mobile phone number *" htmlFor="phone">
                        <Input id="phone" type="tel" value={form.phone} onChange={(e) => update("phone", e.target.value)} placeholder="+7 777 123 4567" required />
                      </Field>
                      <Field label="Telegram *" htmlFor="telegram">
                        <Input id="telegram" value={form.telegram} onChange={(e) => update("telegram", e.target.value)} placeholder="@username" required />
                      </Field>
                      <Field label="Instagram" htmlFor="instagram">
                        <Input id="instagram" value={form.instagram} onChange={(e) => update("instagram", e.target.value)} placeholder="@username" />
                      </Field>
                      <Field label="WhatsApp" htmlFor="whatsapp">
                        <Input id="whatsapp" value={form.whatsapp} onChange={(e) => update("whatsapp", e.target.value)} placeholder="+7 777 123 4567" />
                      </Field>
                    </div>
                  </div>
                </div>
              )}

              {/* ═══ SECTION 3: Education ═══ */}
              {activeSection === "education" && (
                <div className="space-y-6">
                  <h2 className="text-lg font-bold text-gray-900">Education</h2>

                  {/* YouTube */}
                  <div>
                    <h3 className="text-sm font-semibold text-gray-700 mb-1">Personal Presentation</h3>
                    <p className="text-xs text-gray-500 mb-3">Please submit the link to your video presentation.</p>
                    <Field label="YouTube link to your presentation *" htmlFor="youtube_url">
                      <Input id="youtube_url" type="url" value={form.youtube_url} onChange={(e) => update("youtube_url", e.target.value)}
                        placeholder="https://www.youtube.com/watch?v=..." required />
                    </Field>
                  </div>

                  {/* English proficiency */}
                  <div>
                    <h3 className="text-sm font-semibold text-gray-700 mb-1">English proficiency results</h3>
                    <p className="text-xs text-gray-500 mb-3">Please submit the results of your English proficiency test.</p>
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <Field label="Exam *" htmlFor="exam_type">
                        <div className="relative">
                          <select value={form.exam_type} onChange={(e) => update("exam_type", e.target.value)} required
                            className="w-full appearance-none bg-white border border-input rounded-lg px-3 py-2.5 pr-10 text-sm focus:outline-none focus:ring-2 focus:ring-ring">
                            <option value="">Select exam...</option>
                            <option value="IELTS">IELTS</option>
                            <option value="TOEFL">TOEFL</option>
                          </select>
                          <ChevronDown size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none" />
                        </div>
                      </Field>
                      {form.exam_type === "IELTS" && (
                        <Field label="IELTS Score *" htmlFor="ielts_score">
                          <Input id="ielts_score" type="number" step="0.5" min="0" max="9" value={form.ielts_score}
                            onChange={(e) => update("ielts_score", e.target.value)} placeholder="e.g. 7.0" required />
                          <p className="text-xs text-gray-400 mt-1">Minimum 6.0</p>
                        </Field>
                      )}
                      {form.exam_type === "TOEFL" && (
                        <Field label="TOEFL Score *" htmlFor="toefl_score">
                          <Input id="toefl_score" type="number" min="0" max="120" value={form.toefl_score}
                            onChange={(e) => update("toefl_score", e.target.value)} placeholder="e.g. 80" required />
                          <p className="text-xs text-gray-400 mt-1">Minimum 60</p>
                        </Field>
                      )}
                    </div>
                    <div className="mt-3">
                      <FileUploadField label="Copy of your results *" file={englishCertFile}
                        onClear={() => setEnglishCertFile(null)} onClickUpload={() => englishCertRef.current?.click()} />
                      <input ref={englishCertRef} type="file" accept=".jpg,.jpeg,.png,.heic,.pdf" className="hidden"
                        onChange={(e) => handleFileChange(e, setEnglishCertFile)} />
                    </div>
                  </div>

                  {/* Certificate */}
                  <div>
                    <h3 className="text-sm font-semibold text-gray-700 mb-1">Certificate</h3>
                    <div className="grid grid-cols-1 gap-4">
                      <Field label="Certificate type *" htmlFor="certificate_type">
                        <div className="relative">
                          <select value={form.certificate_type} onChange={(e) => update("certificate_type", e.target.value)} required
                            className="w-full appearance-none bg-white border border-input rounded-lg px-3 py-2.5 pr-10 text-sm focus:outline-none focus:ring-2 focus:ring-ring">
                            <option value="">Select type...</option>
                            <option value="UNT">UNT</option>
                            <option value="NIS 12 Grade Certificate">NIS 12 Grade Certificate</option>
                          </select>
                          <ChevronDown size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none" />
                        </div>
                      </Field>
                      <FileUploadField label="Copy of your certificate *" file={certificateFile}
                        onClear={() => setCertificateFile(null)} onClickUpload={() => certificateRef.current?.click()} />
                      <input ref={certificateRef} type="file" accept=".jpg,.jpeg,.png,.heic,.pdf" className="hidden"
                        onChange={(e) => handleFileChange(e, setCertificateFile)} />
                    </div>
                  </div>

                  {/* Additional documents */}
                  <div>
                    <h3 className="text-sm font-semibold text-gray-700 mb-1">Additional documents</h3>
                    <p className="text-xs text-gray-500 mb-3">If you have any additional information about your educational background, you can upload it here.</p>
                    <FileUploadField label="Documents" file={additionalDocsFile}
                      onClear={() => setAdditionalDocsFile(null)} onClickUpload={() => additionalDocsRef.current?.click()} />
                    <input ref={additionalDocsRef} type="file" accept=".pdf,.jpg,.jpeg,.png,.heic" className="hidden"
                      onChange={(e) => handleFileChange(e, setAdditionalDocsFile)} />
                  </div>

                  {/* Essays */}
                  <div className="space-y-4">
                    <Field label="Achievements *" htmlFor="achievements">
                      <Textarea id="achievements" rows={3} value={form.achievements} onChange={(e) => update("achievements", e.target.value)}
                        placeholder="List your key achievements..." required />
                    </Field>
                    <Field label="Extracurricular Activities *" htmlFor="extracurriculars">
                      <Textarea id="extracurriculars" rows={3} value={form.extracurriculars} onChange={(e) => update("extracurriculars", e.target.value)}
                        placeholder="Clubs, volunteer work, sports, hobbies..." required />
                    </Field>
                    <Field label="Essay *" htmlFor="essay">
                      <p className="text-xs text-gray-500 mb-2">Tell us about yourself, your experiences, and why you want to join inVision U.</p>
                      <Textarea id="essay" rows={8} value={form.essay} onChange={(e) => update("essay", e.target.value)}
                        placeholder="Write your essay here..." required />
                      <p className={`text-xs mt-1 ${form.essay.length < 50 && form.essay.length > 0 ? "text-red-500" : "text-gray-400"}`}>
                        {form.essay.length} characters (minimum 50)
                      </p>
                    </Field>
                    <Field label="Motivation Statement *" htmlFor="motivation_statement">
                      <Textarea id="motivation_statement" rows={4} value={form.motivation_statement}
                        onChange={(e) => update("motivation_statement", e.target.value)}
                        placeholder="What motivates you to apply to inVision U?" required />
                    </Field>
                  </div>
                </div>
              )}

              {/* ═══ SECTION 4: Internal Test ═══ */}
              {activeSection === "test" && (
                <div className="space-y-6">
                  <h2 className="text-lg font-bold text-gray-900">Internal Test</h2>
                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <p className="text-sm text-blue-800 font-medium">This is a personality test.</p>
                    <p className="text-xs text-blue-700 mt-1">
                      There are no right or wrong answers — we just want to understand you better, your way of thinking,
                      and what drives your decisions. Be honest and go with your first instinct.
                    </p>
                  </div>

                  <p className="text-sm text-gray-500">Answered: {answeredCount}/40</p>

                  <div className="space-y-6">
                    {PERSONALITY_QUESTIONS.map((pq, qIdx) => (
                      <div key={qIdx} className="border rounded-lg p-4">
                        <p className="text-sm font-medium text-gray-900 mb-3">{pq.q}</p>
                        <div className="space-y-2">
                          {pq.options.map((opt, optIdx) => (
                            <label key={optIdx}
                              className={`flex items-start gap-3 p-2.5 rounded-lg border cursor-pointer transition-colors text-sm ${
                                personalityAnswers[qIdx] === optIdx
                                  ? "border-lime-400 bg-lime-50"
                                  : "border-gray-200 hover:border-gray-300"
                              }`}>
                              <input type="radio" name={`q_${qIdx}`} checked={personalityAnswers[qIdx] === optIdx}
                                onChange={() => setPersonalityAnswers({ ...personalityAnswers, [qIdx]: optIdx })}
                                className="mt-0.5 accent-lime-500" />
                              <span className="text-gray-700">{opt}</span>
                            </label>
                          ))}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          {/* ═══ Static agreements + submit (always visible) ═══ */}
          <Card className="border bg-white mb-6">
            <CardContent className="p-6 space-y-4">
              <label className="flex items-start gap-3 cursor-pointer">
                <input type="checkbox" checked={agreePrivacy} onChange={(e) => setAgreePrivacy(e.target.checked)}
                  className="mt-1 accent-lime-500" />
                <span className="text-sm text-gray-700">
                  By submitting this form, you agree to the processing of your personal data in accordance with our Privacy Policy
                  <span className="text-red-500 ml-0.5">*</span>
                </span>
              </label>

              <label className="flex items-start gap-3 cursor-pointer">
                <input type="checkbox" checked={agreeAge} onChange={(e) => setAgreeAge(e.target.checked)}
                  className="mt-1 accent-lime-500" />
                <span className="text-sm text-gray-700">
                  If the participant is under the age of 18, this questionnaire must be completed by their parent or legal guardian.
                  By proceeding, you confirm that you are either (a) the participant aged 18 or older, or (b) the parent or legal
                  guardian completing this form on behalf of a minor.
                  <span className="text-red-500 ml-0.5">*</span>
                </span>
              </label>

              {error && (
                <p className="text-red-500 text-sm flex items-center gap-1.5">
                  <AlertTriangle size={14} /> {error}
                </p>
              )}

              <Button type="submit" className="w-full py-6 text-base font-semibold" style={{ backgroundColor: "#c1f11d", color: "#111827" }}
                disabled={submitting}>
                {submitting ? "Submitting..." : "Submit Application"}
              </Button>
            </CardContent>
          </Card>
        </form>

        <p className="text-center text-gray-400 text-xs pb-6">
          inVision U Admissions &mdash; Powered by AI screening technology
        </p>
      </div>
    </div>
  );
}

function Field({ label, htmlFor, children }: { label: string; htmlFor: string; children: React.ReactNode }) {
  const hasAsterisk = label.includes("*");
  const labelText = hasAsterisk ? label.replace(" *", "").replace("*", "") : label;
  return (
    <div>
      <Label htmlFor={htmlFor} className="text-sm font-medium text-gray-900 mb-1 block">
        {labelText}{hasAsterisk && <span className="text-red-500 ml-0.5">*</span>}
      </Label>
      {children}
    </div>
  );
}

function FileUploadField({ label, file, onClear, onClickUpload }: {
  label: string; file: File | null; onClear: () => void; onClickUpload: () => void;
}) {
  const hasAsterisk = label.includes("*");
  const labelText = hasAsterisk ? label.replace(" *", "").replace("*", "") : label;
  return (
    <div>
      <p className="text-sm font-medium text-gray-900 mb-1">
        {labelText}{hasAsterisk && <span className="text-red-500 ml-0.5">*</span>}
      </p>
      {file ? (
        <div className="flex items-center gap-3 p-3 border rounded-lg bg-gray-50">
          <span className="text-sm text-gray-700 flex-1 truncate">{file.name}</span>
          <button type="button" onClick={onClear} className="text-red-500 hover:text-red-700">
            <X size={16} />
          </button>
        </div>
      ) : (
        <button type="button" onClick={onClickUpload}
          className="w-full p-4 border-2 border-dashed rounded-lg flex flex-col items-center gap-1 text-gray-400 hover:border-gray-400 hover:text-gray-600 transition-colors">
          <Upload size={20} />
          <span className="text-xs">Click to upload or drag and drop</span>
          <span className="text-[10px] text-gray-400">JPG, JPEG, PNG, HEIC, PDF. Max 10 MB</span>
        </button>
      )}
    </div>
  );
}
