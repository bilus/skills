---
name: prose-contract
description: "Write prose under the strict style contract: outline-review-draft-verify process, actor discipline, banned-phrase and judgment rules with scripted checks."
---

# Prose contract

This contract governs style and procedure. It contains no facts about any document's subject, and its examples are drawn from unrelated domains (cooking, gardening, household machinery) on purpose. Never reuse an example's wording or subject matter in the document. An example phrase appearing in the output is itself a violation (rule C1). The examples illustrate the rules; they are not themselves held to the contract.

The assignment (the prompt this skill is invoked for) supplies the topic, length, framing, output format, and any rule exemptions. Where this skill says "the assignment", read that.

## Trigger

The user asks for a prose deliverable and invokes this skill, mentions the prose contract, or asks for contract-checked writing.

## Actors and actions

Write in terms of actions. Every sentence has a clear actor doing something. Any object can be an actor depending on the context: the oven loses heat, the valve sticks, the letter arrives. Don't forget the author and the reader as actors. Engage the reader with a direct address or two; do not exceed two in the whole document.

If the assignment does not name the main actors, stop before step 1 of the process and ask the user: "Who or what are the main actors in this piece? Name the people involved and the two or three objects whose actions drive it." Wait for the answer, then proceed.

## Steps

1. Verify factual and technical claims against the assignment's named sources first (run code, fetch docs) so the outline rests on checked facts.
2. Write a high-level outline: five to seven claims, one per future paragraph, each a full sentence stating what the paragraph will establish. A paragraph's job is to prove, explain, or clarify its claim. Drop any claim that does not contribute to the assignment's topic.
3. Spawn a reviewer sub-agent with the outline. It answers: does every claim contribute to the topic, is any tension the assignment calls for real, would the assignment's named sources back each factual claim? It returns a numbered list of problems or "none". Fix, then proceed.
4. Expand each claim into two to four sub-claims. Same review.
5. Draft the prose from the expanded outline.
6. Run sections A and C below as a script over the draft, and check the length the assignment sets with wc, excluding whatever the assignment excludes. Fix every hit.
7. Spawn a reviewer sub-agent with the draft and section B only, plus any rule exemptions the assignment declares. It returns each violation as: rule id, quoted sentence, suggested rewrite. Fix, then repeat steps 6 and 7 until the reviewer returns zero violations, at most three rounds.
8. Output only the final document.

## A. Machine-checkable rules

A1. Never use these words or phrases: delve, delves, delving, tapestry, testament, intricate, intricacies, meticulous, meticulously, pivotal, underscore, underscores, underscoring, showcase, showcases, showcasing, commendable, boasts, camaraderie, amidst, realm, interplay, multifaceted, embark, spearhead, garner, labyrinth, symphony, encompassing, unravel, expedite, indelible, captivating, It is important to note, plays a vital role, pave the way, align with, fostering, highlighting, vibrant, elevate, ever-evolving, navigating the complexities, unlock the potential, emphasizing, enhance.

A2. Use at most once per document: crucial, comprehensive, leverage, notably, innovative, utilize, invaluable, noteworthy, robust, foster, navigate, seamless, holistic, landscape, enduring, cutting-edge, game-changer, streamlined, myriad, plethora, unwavering, paramount, groundbreaking, transformative, nuanced, dynamic, powerful, essential, significant, profound, remarkable.

A3. At most two formal transitions per section, from: Furthermore, Moreover, Additionally, Consequently, Nevertheless, Subsequently, Indeed, Hence, Thus, In conclusion, In summary, It's worth noting that, Importantly, Overall, Ultimately, That said, In essence, Therefore, In addition, In contrast, As a result, For instance, For example, In particular, Similarly, Specifically, Crucially, Essentially, Fundamentally, Significantly.

A4. Never "serves as", "stands as", "boasts". Use "is" and "has".
A5. Never "not just X, but Y" or "it's not X, it's Y".
A6. Never "not because X, but because Y".
A7. No "the real question is" or "at its core".
A8. Never "whether you are X or Y".
A9. No stacked "from X to Y, from A to B" ranges.
A10. No "not on the X alone" contrasts.
A11. No "the reason is X" reveals.
A12. No ", both X" appositive tails.
A13. No em dashes or en dashes.
A14. Straight quotes only.
A15. Prefer a period over a semicolon splice.
A16. Lowercase the word after a colon unless it is a proper noun (script flags, author decides).
A17. No bold lead-ins on bullets.
A18. No emoji.
A19. Headings in sentence case.
A20. At most two bold phrases per paragraph.
A21. No one-word rhetorical questions.

## B. Judgment rules

Each rule: statement, then a bad and a good version.

B1. Never end a sentence with a comma and an -ing clause that adds analysis.
    Bad: The dough rose overnight, revealing the yeast's persistence.
    Good: The dough rose overnight.
B2. Do not pad a sentence to three parallel items.
    Bad: The soup needed salt, patience, and a careful hand.
    Good: The soup needed salt.
B3. Attribute a claim to a named source or drop it.
    Bad: Experts agree tomatoes dislike cold soil.
    Good: Cornell's vegetable guide says to plant tomatoes after the soil holds 60F. (Or delete the claim.)
B4. Never inflate importance; state the consequence or say nothing.
    Bad: Watering matters more than most people realize.
    Good: Skip a week of watering and the seedlings die.
B5. No asides addressed to the reader. (The direct addresses permitted under "Actors and actions" do work; an aside does not.)
    Bad: The compost heats up fast (you have probably noticed this too).
    Good: The compost heats up fast.
B6. Every number needs a named source. Numbers native to the document's own narrative (dates, counts the narrator observed) count as sourced by the narrative; the assignment may declare further exemptions.
    Bad: Most gardeners overwater.
    Good: The extension office's survey found overwatering in six of ten plots it visited. (Or drop it.)
B7. At most one hedge per sentence; otherwise make the claim.
    Bad: It might perhaps help to possibly soak the beans.
    Good: Soaking the beans may help.
B8. A list item must say something its label does not.
    Bad: - Watering: this is about watering the plants.
    Good: - Watering: once at dawn, skipped on rain days.
B9. Never announce the content; state it.
    Bad: Let me explain why the compost heats up.
    Good: The compost heats up because the bacteria work fastest at the pile's wet center.
B10. One punchy fragment is fine; a drumroll of them is not.
    Bad: No rain. No shade. No hope.
    Good: No rain for three weeks, and the shade cloth was still in the garage.
B11. No mid-sentence re-explanations of the subject.
    Bad: The starter, which is just flour and water left to ferment, needed feeding. (When the starter was already defined.)
    Good: The starter needed feeding.
B12. No asides confirming your own claim.
    Bad: The frost killed the basil, as frost does.
    Good: The frost killed the basil.
B13. No generic optimism.
    Bad: With a little care, any garden can thrive.
    Good: (delete)
B14. No outlet parades; one dated, specific citation.
    Bad: As reported by the almanac, the co-op newsletter, and several forums...
    Good: The 2024 county almanac lists the last frost as May 12.
B15. No manufactured aphorisms.
    Bad: A watched kettle teaches patience.
    Good: (delete)
B16. Never describe code by its diff history; describe what it is.
    Bad: The backup script no longer retries on failure.
    Good: The backup script runs once and reports failure.
B17. Answer a question with the answer, in the first sentence.
    Bad: Why did the loaf collapse? There are several factors worth examining.
    Good: The loaf collapsed because the oven lost heat when the door opened.
B18. End a section on its last fact; no summary or closing-optimism sentence.
    Bad: ...and the hinge stopped squeaking. In the end, small fixes like this are what home ownership is all about.
    Good: ...and the hinge stopped squeaking.
B19. No orphan mini-paragraph under a heading.
B20. Negate the verb, not the object.
    Bad: The recipe allows no substitutions.
    Good: The recipe doesn't allow substitutions.
B21. No stilted connectives.
    Bad: The dough, thusly rested, was ready.
    Good: After resting, the dough was ready.
B22. No travel-brochure puffery.
    Bad: the sun-drenched rows of the allotment
    Good: the allotment's south rows
B23. No figurative landing for software changes.
    Bad: With the fix in, the nightly job finally breathed.
    Good: With the fix in, the nightly job finished.
B24. No counted lead-ins announcing what follows.
    Bad: Three things went wrong that morning.
    Good: The freezer door was open. (Then the next thing.)
B25. No mirrored-suffix contrasts.
    Bad: The problem was not durability but flexibility.
    Good: The handle flexed under load.
B26. No infallibility claims for machinery.
    Bad: The thermostat never misses.
    Good: The thermostat reads the sensor once a minute.
B27. No knowing flourishes for deletions.
    Bad: I deleted the macro. It will not be missed.
    Good: I deleted the macro.
B28. No in-and-out mirror cadences.
    Bad: You go in with questions and come out with better questions.
    Good: (state what was learned, or delete)

## C. Leak check

C1. No phrase of four or more consecutive words from this skill's examples may appear in the document. Check with a script.

## Verification

The document passes when: the section A and C script reports zero hits; the word count sits inside the assignment's range, excluding what the assignment excludes; the section B reviewer sub-agent returns zero violations (within three rounds); direct reader addresses number at most two; and every factual claim traces to the assignment's named source, which the document names once when the assignment says so.