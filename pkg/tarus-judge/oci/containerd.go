package oci_judge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	hr_bytes "github.com/Myriad-Dreamin/tarus/pkg/hr-bytes"
	tarus_io "github.com/Myriad-Dreamin/tarus/pkg/tarus-io"
	tarus_judge "github.com/Myriad-Dreamin/tarus/pkg/tarus-judge"
	tarus_store "github.com/Myriad-Dreamin/tarus/pkg/tarus-store"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	v1 "github.com/containerd/containerd/metrics/types/v1"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/typeurl"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ContainerdJudgeOption func(svc *ContainerdJudgeServiceServer) error

type ContainerdJudgeConfig struct {
	Address        string `json:"address"`
	JudgeCachePath string `json:"judge_cache_path"`
	Concurrency    int    `json:"concurrency"`
	JudgeWorkdir   string `json:"judge_workdir"`
}

type ContainerdJudgeServiceServer struct {
	tarus.UnimplementedJudgeServiceServer
	client       *containerd.Client
	sessionStore tarus_store.JudgeSessionStore
	closers      []io.Closer
	ioRouter     tarus_io.Router

	options   ContainerdJudgeConfig
	ccLimiter chan int
}

func WithContainerdAddress(address string) ContainerdJudgeOption {
	return func(svc *ContainerdJudgeServiceServer) error {
		svc.options.Address = address
		return nil
	}
}

func WithContainerdJudgeCachePath(path string) ContainerdJudgeOption {
	return func(svc *ContainerdJudgeServiceServer) error {
		svc.options.JudgeCachePath = path
		return nil
	}
}

func WithContainerdConcurrencyNum(cc int) ContainerdJudgeOption {
	return func(svc *ContainerdJudgeServiceServer) error {
		svc.options.Concurrency = cc
		return nil
	}
}

func WithContainerdJudgeWorkdir(wd string) ContainerdJudgeOption {
	return func(svc *ContainerdJudgeServiceServer) error {
		svc.options.JudgeWorkdir = wd
		return nil
	}
}

func defaultContainerdJudgeConfig() ContainerdJudgeConfig {
	return ContainerdJudgeConfig{
		Address:        "/run/containerd/containerd.sock",
		JudgeCachePath: "./test.db",
		JudgeWorkdir:   "./data/workdir-judge-engine{cid}",
		Concurrency:    1,
	}
}

func NewContainerdServer(options ...ContainerdJudgeOption) (svc *ContainerdJudgeServiceServer, err error) {
	svc = &ContainerdJudgeServiceServer{
		ioRouter: tarus_io.Statics,
		options:  defaultContainerdJudgeConfig(),
	}

	for i := range options {
		if err := options[i](svc); err != nil {
			return nil, err
		}
	}

	if svc.options.Concurrency <= 0 {
		return nil, fmt.Errorf("invalid conccurency option, should be greater than zero: %d", svc.options.Concurrency)
	}
	svc.ccLimiter = make(chan int, svc.options.Concurrency)
	for i := 0; i < svc.options.Concurrency; i++ {
		svc.ccLimiter <- i
	}

	svc.client, err = containerd.New(svc.options.Address,
		containerd.WithDefaultNamespace("tarus"))
	if err != nil {
		return
	}

	svc.closers = append(svc.closers, svc.client)
	defer func() {
		if err != nil {
			_ = svc.Close()
		}
	}()

	b, err := bbolt.Open(svc.options.JudgeCachePath, os.FileMode(0644), nil)
	if err != nil {
		return
	}
	svc.sessionStore = tarus_store.NewJudgeSessionStore(tarus_store.NewDB(b))
	if svc.sessionStore == nil {
		err = errors.Wrapf(errdefs.ErrInvalidArgument, "session store not filled")
		return
	}

	if c, ok := svc.sessionStore.(io.Closer); ok {
		svc.closers = append(svc.closers, c)
	}
	return
}

func (c *ContainerdJudgeServiceServer) Close() error {
	for i := range c.closers {
		_ = c.closers[i].Close()
	}
	return nil
}

func (c *ContainerdJudgeServiceServer) Handshake(_ context.Context, request *tarus.HandshakeRequest) (*tarus.HandshakeResponse, error) {
	if !bytes.Equal(request.ApiVersion, []byte("v0.0.0")) {
		return nil, status.Error(codes.FailedPrecondition, "client version not handled by service")
	}

	return &tarus.HandshakeResponse{
		ApiVersion:      ContainerdJudgeVersion,
		JudgeStatusHash: tarus_judge.JudgeStatusHash,
		ImplementedApis: []string{
			tarus_judge.JudgeServiceApiMinimum,
		},
	}, nil
}

func (c *ContainerdJudgeServiceServer) CopyFile(ctx context.Context, request *tarus.CopyRequest) (*emptypb.Empty, error) {
	return c.UnimplementedJudgeServiceServer.CopyFile(ctx, request)
}

func (c *ContainerdJudgeServiceServer) CreateContainer(ctx context.Context, request *tarus.CreateContainerRequest) (_ *emptypb.Empty, err error) {
	ctx = namespaces.WithNamespace(ctx, "tarus")

	// prepare image
	snapshotter := containerd.DefaultSnapshotter
	if err = c.prepareImageOnSnapshotter(ctx, request.ImageId, snapshotter); err != nil {
		return nil, err
	}
	image, err := c.client.GetImage(ctx, request.ImageId)
	if err != nil {
		return nil, err
	}

	// generate config used by worker
	ctrId := <-c.ccLimiter
	fixedContainerId := fmt.Sprintf("tarus-engine-snapshot%d", ctrId)
	fixedContainerSnapshotId := fmt.Sprintf("tarus-engine-snapshot%d", ctrId)
	fixedWorkDir := strings.ReplaceAll(c.options.JudgeWorkdir, "{cid}", strconv.Itoa(ctrId))

	// clear container in daemon
	if err = c.client.SnapshotService(snapshotter).Remove(ctx, fixedContainerSnapshotId); err != nil && !errdefs.IsNotFound(err) {
		return nil, err
	}
	if err = c.client.ContainerService().Delete(ctx, fixedContainerId); err != nil && !errdefs.IsNotFound(err) {
		return nil, err
	}

	// prepare mount options
	mounts := make([]specs.Mount, 0)
	m, err := filepath.Abs(fixedWorkDir)
	if err != nil {
		return nil, err
	}
	mounts = append(mounts, specs.Mount{
		// notes: linux only flags
		Type:        "bind",
		Source:      m,
		Destination: "/workdir",
		Options:     []string{"rbind", "ro"},
	})

	container, err := c.client.NewContainer(
		ctx,
		fixedContainerId,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(fixedContainerSnapshotId, image),
		containerd.WithNewSpec(
			oci.WithImageConfig(image),
			oci.WithMounts(mounts),
			oci.WithProcessCwd("/workdir"),
		),
		// containerd.WithRuntime("/home/kamiyoru/work/go/tarus/cmd/runsc/wrapper.template.sh", nil),
	)

	if err != nil {
		return nil, err
	}
	if container == nil {
		return nil, status.Error(codes.Internal, "container not created when successfully returning")
	}

	var session = tarus.OCIJudgeSession{
		CommitStatus: 0,
		ContainerId:  fixedContainerId,
		HostWorkdir:  fixedWorkDir,
		WorkerId:     int32(ctrId),
		BinTarget:    request.GetBinTarget(),
	}
	err = c.sessionStore.SetJudgeSession(ctx, request.TaskKey, &session)
	if err != nil {
		return nil, err
	}

	return new(emptypb.Empty), nil
}

func (c *ContainerdJudgeServiceServer) RemoveContainer(ctx context.Context, request *tarus.RemoveContainerRequest) (*emptypb.Empty, error) {
	ctx = namespaces.WithNamespace(ctx, "tarus")

	// todo, detect workload atomic

	session, err := c.sessionStore.GetJudgeSession(ctx, request.TaskKey)
	if err != nil {
		return nil, err
	}

	c.ccLimiter <- int(session.WorkerId)

	cc, err := c.client.LoadContainer(ctx, session.ContainerId)
	if err != nil {
		return nil, err
	}

	err = c.sessionStore.FinishSession(ctx, request.TaskKey, func() error {
		return cc.Delete(ctx, containerd.WithSnapshotCleanup)
	})
	if err != nil {
		return nil, err
	}

	return new(emptypb.Empty), nil
}

func getOrDefault(L, R int64) int64 {
	if L != 0 {
		return L
	}
	return R
}

func (c *ContainerdJudgeServiceServer) MakeJudge(rawCtx context.Context, request *tarus.MakeJudgeRequest) (*tarus.MakeJudgeResponse, error) {
	ctx := namespaces.WithNamespace(rawCtx, "tarus")

	session, err := c.sessionStore.GetJudgeSession(ctx, request.TaskKey)
	if err != nil {
		return nil, err
	}

	cc, err := c.client.LoadContainer(ctx, session.ContainerId)
	if err != nil {
		return nil, err
	}

	// todo: config cgroup
	s, err := cc.Spec(ctx)
	if err != nil {
		return nil, err
	}
	procTmpl := *s.Process
	reqLevelCpuhard := getOrDefault(request.Cpuhard, int64(15*time.Second))
	reqLevelCputime := getOrDefault(request.Cputime, int64(10*time.Second))
	reqLevelMemory := getOrDefault(request.Memory, int64(1024*hr_bytes.MB))
	reqLevelStack := getOrDefault(request.Stack, int64(1024*hr_bytes.MB))

	var resp = new(tarus.MakeJudgeResponse)
	for i := range request.Testcases {
		err = c.withFreshTask(ctx, cc, func(t containerd.Task) error {
			// fmt.Printf("linux container create successfully\n")

			var judgePoint = request.Testcases[i]
			procOpts := procTmpl
			procOpts.Terminal = false
			procOpts.Args = []string{session.BinTarget}

			var ioProvider = judgePoint.IoProvider
			if len(ioProvider) == 0 {
				ioProvider = request.IoProvider
			}

			ioc, err := c.ioRouter.MakeIOChannel(ioProvider)
			if err != nil {
				return err
			}

			fac, err := ioc(judgePoint.Input, judgePoint.Answer)
			if err != nil {
				return err
			}

			cpuhard := getOrDefault(judgePoint.EstimatedCpuhard, reqLevelCpuhard)
			cputime := getOrDefault(judgePoint.EstimatedCputime, reqLevelCputime)
			memory := getOrDefault(judgePoint.EstimatedMemory, reqLevelMemory)
			stack := getOrDefault(judgePoint.EstimatedStack, reqLevelStack)

			// todo: check default rlimit, check rlimit_core
			if len(procOpts.Rlimits) != 0 {
				procOpts.Rlimits = procOpts.Rlimits[:0]
			}

			if cputime > 1 || cpuhard > 1 {
				if cputime > 1 && cpuhard > 1 {
					procOpts.Rlimits = append(procOpts.Rlimits, specs.POSIXRlimit{
						Type: "RLIMIT_CPU",
						Soft: uint64(cputime),
						Hard: uint64(cpuhard),
					})
				} else {
					return errors.New("both cpu hard and cpu time should be set at the same time")
				}
			}

			if memory > 1 {
				procOpts.Rlimits = append(procOpts.Rlimits, specs.POSIXRlimit{
					Type: "RLIMIT_DATA",
					Soft: uint64(memory),
					Hard: uint64(memory + (32 * int64(hr_bytes.MB))),
				})
			}

			if stack > 1 {
				procOpts.Rlimits = append(procOpts.Rlimits, specs.POSIXRlimit{
					Type: "RLIMIT_STACK",
					Soft: uint64(stack),
					Hard: uint64(stack + (32 * int64(hr_bytes.MB))),
				})
			}

			err = (func() error {
				process, err := t.Exec(ctx, "judge_exec", &procOpts, fac.AsCreator())
				if err != nil {
					return err
				}
				defer func() {
					_, _ = process.Delete(ctx)
				}()

				statusC, err := process.Wait(ctx)
				if err != nil {
					return err
				}

				var processStartTime = time.Now()
				if err := process.Start(ctx); err != nil {
					return err
				}
				defer func() {
					// todo: check process status
					_ = c.killProcess(ctx, process)
				}()

				select {
				case st := <-statusC:
					var qr = &tarus.QueryJudgeItem{
						JudgeKey: judgePoint.JudgeKey,
					}
					var jh JudgeHint
					code, exitedAt, err := st.Result()
					if err != nil {
						return err
					}
					jh.Code = int(code)
					qr.TimeUseHard = int64(exitedAt.Sub(processStartTime) * time.Nanosecond)

					s, err := fac.GetJudgeResult()
					if err != nil {
						return err
					}
					jh.CheckerResult = string(s)

					m, err := t.Metrics(rawCtx)
					if err != nil {
						return err
					}

					m0, err := typeurl.UnmarshalAny(m.Data)
					if err != nil {
						return err
					}
					if m2, ok := m0.(*v1.Metrics); ok {
						//fmt.Printf("metrics cpu result: %v %v %v\n", m2.CPU.Usage.Total, m2.CPU.Usage.Kernel, m2.CPU.Usage.User)
						//fmt.Printf("metrics memory result: %v %v %v\n", m2.Memory.Usage.Max, m2.Memory.Usage.Usage, m2.Memory.RSS)
						qr.TimeUse = int64(m2.CPU.Usage.User)
						qr.MemoryUse = int64(m2.Memory.Usage.Max)
					} else {
						fmt.Println("invalid type url for extracting metrics", m.Data.TypeUrl)
					}

					qr.Hint, err = json.Marshal(jh)
					if err != nil {
						return err
					}

					qr.Status, err = fac.GetJudgeStatus(s)
					if err != nil {
						return err
					}
					resp.Items = append(resp.Items, qr)
				case <-time.After(time.Duration(cpuhard) * time.Nanosecond):
					fmt.Printf("linux container timeout stop\n")
				}

				return nil
			})()
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func (c *ContainerdJudgeServiceServer) QueryJudge(ctx context.Context, request *tarus.QueryJudgeRequest) (*tarus.QueryJudgeResponse, error) {
	return c.UnimplementedJudgeServiceServer.QueryJudge(ctx, request)
}

func (c *ContainerdJudgeServiceServer) ImportOCIArchiveR(ctx context.Context, f io.Reader, ref string) error {
	ctx = namespaces.WithNamespace(ctx, "tarus")

	_, err := c.client.Import(ctx, f, containerd.WithImageRefTranslator(func(s string) string {
		fmt.Println("s", s)
		if len(ref) != 0 {
			return ref
		}

		return s
	}))
	if err != nil {
		return err
	}

	return nil
}

func (c *ContainerdJudgeServiceServer) ImportOCIArchive(ctx context.Context, fp string) error {
	ctx = namespaces.WithNamespace(ctx, "tarus")

	f, err := os.OpenFile(fp, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	_, err = c.client.Import(ctx, f)
	_ = f.Close()
	if err != nil {
		return err
	}

	return nil
}
