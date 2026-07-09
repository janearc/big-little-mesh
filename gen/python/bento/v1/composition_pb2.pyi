from bento.v1 import bento_pb2 as _bento_pb2
from frood.v1 import frood_pb2 as _frood_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RecipeStep(_message.Message):
    __slots__ = ("capability", "input", "output")
    CAPABILITY_FIELD_NUMBER: _ClassVar[int]
    INPUT_FIELD_NUMBER: _ClassVar[int]
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    capability: str
    input: _frood_pb2.ContractRef
    output: _frood_pb2.ContractRef
    def __init__(self, capability: _Optional[str] = ..., input: _Optional[_Union[_frood_pb2.ContractRef, _Mapping]] = ..., output: _Optional[_Union[_frood_pb2.ContractRef, _Mapping]] = ...) -> None: ...

class Convergence(_message.Message):
    __slots__ = ("acceptance", "max_passes")
    ACCEPTANCE_FIELD_NUMBER: _ClassVar[int]
    MAX_PASSES_FIELD_NUMBER: _ClassVar[int]
    acceptance: _bento_pb2.BanchanAsset
    max_passes: int
    def __init__(self, acceptance: _Optional[_Union[_bento_pb2.BanchanAsset, _Mapping]] = ..., max_passes: _Optional[int] = ...) -> None: ...

class Recipe(_message.Message):
    __slots__ = ("steps", "convergence")
    STEPS_FIELD_NUMBER: _ClassVar[int]
    CONVERGENCE_FIELD_NUMBER: _ClassVar[int]
    steps: _containers.RepeatedCompositeFieldContainer[RecipeStep]
    convergence: Convergence
    def __init__(self, steps: _Optional[_Iterable[_Union[RecipeStep, _Mapping]]] = ..., convergence: _Optional[_Union[Convergence, _Mapping]] = ...) -> None: ...

class Intent(_message.Message):
    __slots__ = ("intent_id", "parent_intent_id", "recipe_id", "bento_id", "submitter", "hop_bound")
    INTENT_ID_FIELD_NUMBER: _ClassVar[int]
    PARENT_INTENT_ID_FIELD_NUMBER: _ClassVar[int]
    RECIPE_ID_FIELD_NUMBER: _ClassVar[int]
    BENTO_ID_FIELD_NUMBER: _ClassVar[int]
    SUBMITTER_FIELD_NUMBER: _ClassVar[int]
    HOP_BOUND_FIELD_NUMBER: _ClassVar[int]
    intent_id: str
    parent_intent_id: str
    recipe_id: str
    bento_id: str
    submitter: _frood_pb2.Identity
    hop_bound: int
    def __init__(self, intent_id: _Optional[str] = ..., parent_intent_id: _Optional[str] = ..., recipe_id: _Optional[str] = ..., bento_id: _Optional[str] = ..., submitter: _Optional[_Union[_frood_pb2.Identity, _Mapping]] = ..., hop_bound: _Optional[int] = ...) -> None: ...
