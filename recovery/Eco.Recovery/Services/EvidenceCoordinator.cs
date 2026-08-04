using Eco.Recovery.Models;

namespace Eco.Recovery.Services;

public sealed class EvidenceCoordinator
{
    private readonly MatterRepository _repository;
    private readonly EvidenceIntakeService _intake;

    public EvidenceCoordinator(MatterRepository repository, EvidenceIntakeService intake)
    {
        _repository = repository ?? throw new ArgumentNullException(nameof(repository));
        _intake = intake ?? throw new ArgumentNullException(nameof(intake));
    }

    public async Task<EvidenceRecord> ImportAsync(
        string matterId,
        string sourcePath,
        IProgress<EvidenceImportProgress>? progress = null,
        CancellationToken cancellationToken = default)
    {
        var workspace = await _repository.LoadAsync(matterId, cancellationToken)
            ?? throw new DirectoryNotFoundException("The requested Matter does not exist.");
        var expectedRevision = workspace.Revision;
        var record = await _intake.PreserveAndExtractAsync(
            sourcePath,
            workspace,
            progress,
            cancellationToken);

        workspace.Evidence.Add(record);
        try
        {
            await _repository.SaveAsync(workspace, expectedRevision, cancellationToken);
            return record;
        }
        catch
        {
            workspace.Evidence.Remove(record);
            _intake.RollBackPreservedEvidence(matterId, record);
            throw;
        }
    }
}
