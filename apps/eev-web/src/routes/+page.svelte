<script lang="ts" module>
    import {z} from "zod/v4";

    const formSchema = z.object({
        code: z.string().min(6, "Share code must be at least 6 characters long")
    })
</script>

<script lang="ts">
    import * as Card from "$lib/components/ui/card/index.js";
    import * as Form from "$lib/components/ui/form/index.js";
    import * as InputOTP from "$lib/components/ui/input-otp/index.js";
    import {Search} from "@lucide/svelte";
    import {defaults, superForm} from "sveltekit-superforms";
    import {zod4} from "sveltekit-superforms/adapters";
    import {Button} from "$lib/components/ui/button";

    const form = superForm(defaults(zod4(formSchema)), {
        validators: zod4(formSchema),
        SPA: true,
        onChange: (event) => {
            const code = event.get('code');
            event.set('code', code.toUpperCase());
        }
    })

    const { form: formData, errors, submitting, enhance } = form;
    $: isSubmitDisabled = Object.keys($errors).length > 0 ||
        !$formData.code ||
        $formData.code.length !== 6 ||
        $submitting;

    $: isResetDisabled = !$formData.code
        || $formData.code.length === 0
        || $submitting;

    function search(){

    }

    function reset(){
        form.reset();
    }


</script>

<div class="min-h-screen flex items-center justify-center ">
    <Card.Root class="w-100 h-72 shadow-lg">
        <Card.Header>
            <Card.Title>Your Share Code</Card.Title>
            <Card.Description>Please enter your share code that you received from the sender</Card.Description>
            <Card.Action>
                <Button variant="link" href="/share" aria-label="share new">Share New</Button>
            </Card.Action>
        </Card.Header>

        <div class="mt-2">
            <Card.Content>
                <form use:enhance method="POST">
                    <Form.Field {form} name="code">
                        <Form.Control>

                            {#snippet children({ props })}
                                <div class="w-full">
                                    <InputOTP.Root class="flex justify-center items-center" maxlength={6} {...props} bind:value={$formData.code}>
                                        {#snippet children({ cells })}
                                            {#each cells.slice(0, 6) as cell (cell)}
                                                <InputOTP.Group>
                                                    <InputOTP.Slot {cell} />
                                                </InputOTP.Group>
                                            {/each}
                                        {/snippet}
                                    </InputOTP.Root>
                                </div>
                            {/snippet}
                        </Form.Control>
                    </Form.Field>
                </form>
            </Card.Content>
        </div>

        <Card.Footer class="flex-col gap-2">
            <Button type="submit" variant="default" class="w-full" aria-label="submit" disabled={isSubmitDisabled} onclick="{() => search()}">
                <Search />
                Search
            </Button>
            <Button variant="destructive" class="w-full" aria-label="reset" disabled={isResetDisabled} onclick="{() => reset()}">Reset</Button>
        </Card.Footer>
    </Card.Root>
</div>